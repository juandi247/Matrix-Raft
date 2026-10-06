const SimulatorEventType = Object.freeze({
  TickAdvance: 0,
  NodeCrashed: 1,
  NodeBackToLife: 2,
  HeartbeatTimeout: 3,
  NodeStateUpdate: 4,
  NewMessage: 5,
  MessageDelivered: 6,
});

/*---------- GOLANG CODE  FOR GUIDE -----------
	type SimulatorEventType int
const (
	TickAdvance SimulatorEventType =  iota
	NodeCrashed 
	NodeBackToLife 	
	HeartbeatTimeout
	NodeStateUpdate
	NewMessage
	MessageDelivered 
)
 */


const SimulatorEventName = Object.freeze({
  [SimulatorEventType.TickAdvance]: 'TickAdvance',
  [SimulatorEventType.NodeCrashed]: 'NodeCrashed',
  [SimulatorEventType.NodeBackToLife]: 'NodeBackToLife',
  [SimulatorEventType.HeartbeatTimeout]: 'HeartbeatTimeout',
  [SimulatorEventType.NodeStateUpdate]: 'NodeStateUpdate',
  [SimulatorEventType.NewMessage]: 'NewMessage',
  [SimulatorEventType.MessageDelivered]: 'MessageDelivered',
});



function handleSimulatorEvent(event) {
  switch (event.EventType) {
    case SimulatorEventType.TickAdvance:
      $('tick').textContent = event.Payload.Tick;
      onTick(event.Payload.Tick);               // remember the tick and the real time between ticks
      deliverDueMessages(event.Payload.Tick);   // arrivals first
      movePackets(event.Payload.Tick);          // then every message in flight glides on to the next tick
      break;

    case SimulatorEventType.NodeCrashed:
      console.log('NodeCrashed event:', event.Payload);
      break;

    case SimulatorEventType.NodeBackToLife:
      console.log('NodeBackToLife event:', event.Payload);
      break;

    case SimulatorEventType.HeartbeatTimeout:
      handleHeartbeatEvent(event.Payload);
      break;

    case SimulatorEventType.NodeStateUpdate:
      renderNodeState(event.Payload);
      break;

    case SimulatorEventType.NewMessage:
      handleNewMessage(event.Payload);
      break;

    case SimulatorEventType.MessageDelivered:
      removeMessageById(event.Payload.MessageId ?? event.Payload.Id);
      break;

    default:
      console.warn('Unknown event type:', event.EventType);
  }
}





function handleHeartbeatEvent(payload) {
  const view = nodeViews.get(payload.NodeId);

  if (!view) {
    console.warn('Heartbeat received for unknown node:', payload.NodeId);
    return;
  }

  view.fields.hb.textContent = payload.HeartbeatTimeoutCounter;
}


//TODO: change this into some other js file
//
// ======================= NODES =======================
// Renders the state of each Raft node inside #nodes.
// Call renderNodeState(payload) for every NodeStateUpdate event.
// The DOM is only touched when a value actually changed, so repeated
// identical updates do nothing (no flicker).

const ROLE_NAMES = ['follower', 'candidate', 'leader']; // Go iota: 0, 1, 2
const ROLE_LEADER = 2;

// Node1, Node2, ..., Node10 in natural order. The maps always show in the same order.
const byNodeName = (a, b) => a.localeCompare(b, undefined, { numeric: true });

const nodeViews = new Map(); // NodeId -> { el, fields, last, log, nextIndex, matchIndex }

// Nodes sit on the corners of a regular polygon (triangle for 3, pentagon for 5, ...).
// Positions are percentages of the .stage: center (cx, cy) and radii (rx, ry).
const STAGE = { cx: 50, cy: 44, rx: 38, ry: 36 };

function layoutNodes() {
  const cards = [...document.getElementById('nodes').children];
  const n = cards.length;
  const start = n === 2 ? Math.PI : -Math.PI / 2; // 2 nodes: side by side. Otherwise the first node is on top
  cards.forEach((el, i) => {
    const angle = start + (i * 2 * Math.PI) / n; // clockwise
    const r = n === 1 ? 0 : 1;
    // (left, top) is the center of the circle, handy for message animations
    el.style.left = STAGE.cx + STAGE.rx * r * Math.cos(angle) + '%';
    el.style.top = STAGE.cy + STAGE.ry * r * Math.sin(angle) + '%';
  });
  drawNodeLinks(); // lines between every pair of nodes
}

function createNodeView(id) {
  const el = document.createElement('article');
  el.className = 'node';
  el.dataset.id = id;
  el.dataset.role = '0'; // every node starts as follower
  el.innerHTML = `
    <div class="node__circle"><span data-f="id"></span></div>
    <div class="node__info">
      <p class="node__role" data-f="role"></p>
      <dl class="node__stats">
        <div><dt>term</dt><dd data-f="term"></dd></div>
        <div><dt>commit</dt><dd data-f="commit"></dd></div>
        <div><dt title="heartbeat timeout">timeout</dt><dd data-f="hb"></dd></div>
      </dl>
      <div class="node__log">
        <p class="node__last" data-f="last">log empty</p>
        <button class="node__more" type="button" data-f="more">full log</button>
        <button class="node__more node__indexes" type="button" data-f="indexes">next / match index</button>
      </div>
    </div>`;

  const fields = {};
  el.querySelectorAll('[data-f]').forEach((f) => { fields[f.dataset.f] = f; });
  fields.id.textContent = id;
  fields.more.addEventListener('click', () => openLogPanel(id));
  fields.indexes.addEventListener('click', () => openIndexPanel(id)); // only visible for the leader (CSS)

  // keep nodes sorted: Node1, Node2, ..., Node10
  const container = document.getElementById('nodes');
  const next = [...container.children].find(
    (c) => c.dataset.id.localeCompare(id, undefined, { numeric: true }) > 0
  );
  container.insertBefore(el, next || null);
  layoutNodes(); // a new node changes the polygon

  return { el, fields, last: {}, log: [], nextIndex: {}, matchIndex: {} };
}

// Runs apply(value) only if the value differs from the last one rendered.
function updateIfChanged(view, key, value, apply) {
  if (view.last[key] === value) return;
  view.last[key] = value;
  apply(value);
}

function formatValue(v) {
  if (v === null || v === undefined) return '';
  return typeof v === 'object' ? JSON.stringify(v) : String(v);
}

// The card only shows the last entry (index starts at 0). The full log is in the panel.
function renderLogSummary(view, log) {
  const { last, more } = view.fields;
  if (log.length === 0) {
    last.textContent = 'log empty';
    more.textContent = 'full log';
    return;
  }
  const i = log.length - 1;
  last.textContent = `last [${i}] · term ${formatValue(log[i].Term ?? log[i].term)}`;
  more.textContent = `full log · ${log.length}`;
}

// payload: { NodeId, Term, Role, Log, CommitIndex, SimulatorHeartBeatTimeoutCounter, NextIndex, MatchIndex }
function renderNodeState(s) {
  let view = nodeViews.get(s.NodeId);
  if (!view) {
    view = createNodeView(s.NodeId);
    nodeViews.set(s.NodeId, view);
  }
  const f = view.fields;

  updateIfChanged(view, 'role', s.Role, (role) => {
    view.el.dataset.role = role; // the CSS shows the "next / match index" button only when role = 2
    f.role.textContent = ROLE_NAMES[role] ?? 'unknown';
    if (role !== ROLE_LEADER && openIndexId === s.NodeId) closeIndexPanel(); // no longer the leader
  });
  updateIfChanged(view, 'term', s.Term, (v) => { f.term.textContent = v; });
  updateIfChanged(view, 'commit', s.CommitIndex, (v) => { f.commit.textContent = v; });
  updateIfChanged(view, 'hb', s.SimulatorHeartBeatTimeoutCounter, (v) => { f.hb.textContent = v; });

  // NextIndex / MatchIndex: maps FollowerId -> index. Go sends null for a nil map.
  const next = s.NextIndex || {};
  const match = s.MatchIndex || {};
  view.nextIndex = next;
  view.matchIndex = match;
  const ids = [...new Set([...Object.keys(next), ...Object.keys(match)])].sort(byNodeName);
  const indexSignature = ids.map((id) => `${id}:${next[id]}:${match[id]}`).join('|');
  updateIfChanged(view, 'indexes', indexSignature, () => {
    if (openIndexId === s.NodeId) renderIndexPanel(next, match); // keep the open panel live
  });

  // The log can be huge, so the change check uses the length and the last entry only.
  const log = s.Log || []; // Go sends null for an empty slice
  view.log = log;
  const signature = `${log.length}|${JSON.stringify(log[log.length - 1] ?? null)}`;
  updateIfChanged(view, 'log', signature, () => {
    renderLogSummary(view, log);
    if (openLogId === s.NodeId) renderLogPanel(log); // keep the open panel live
  });
}


// ======================= FULL LOG PANEL =======================
// Floats above the page. The simulation keeps running behind it.

let openLogId = null;    // NodeId of the open panel, or null
let panelCount = 0;      // rows already rendered in the panel
let panelLastKey = '';   // JSON of the last rendered entry

function logRow(entry, i) {
  const tr = document.createElement('tr');
  tr.append(
    netEl('td', '', i),
    netEl('td', '', formatValue(entry.Term ?? entry.term)),
    netEl('td', '', formatValue(entry.Value ?? entry.value))
  );
  return tr;
}

// Appends only the new entries when the log just grew. Rebuilds otherwise.
function renderLogPanel(log) {
  const body = netById('fulllog-body');
  const scroll = netById('fulllog-scroll');
  const follow = scroll.scrollTop + scroll.clientHeight >= scroll.scrollHeight - 8; // user is at the bottom
  const append = panelCount > 0 && log.length >= panelCount &&
    JSON.stringify(log[panelCount - 1]) === panelLastKey;

  if (!append) { body.replaceChildren(); panelCount = 0; }

  if (log.length === 0) {
    const td = body.insertRow().insertCell();
    td.colSpan = 3;
    td.className = 'muted';
    td.textContent = 'empty';
  } else {
    const rows = document.createDocumentFragment();
    for (let i = panelCount; i < log.length; i++) rows.append(logRow(log[i], i)); // index starts at 0
    body.append(rows);
    panelCount = log.length;
    panelLastKey = JSON.stringify(log[log.length - 1]);
  }

  netById('fulllog-count').textContent = `${log.length} ${log.length === 1 ? 'entry' : 'entries'}`;
  if (follow) scroll.scrollTop = scroll.scrollHeight;
}

function openLogPanel(id) {
  const view = nodeViews.get(id);
  if (!view) return;
  closeIndexPanel(); // only one floating panel at a time
  openLogId = id;
  panelCount = 0;
  netById('fulllog-title').textContent = `${id} · log`;
  netById('fulllog').hidden = false;
  renderLogPanel(view.log);
  netById('fulllog-close').focus();
}

function closeLogPanel() {
  openLogId = null;
  netById('fulllog').hidden = true;
}


// ======================= NEXT / MATCH INDEX PANEL (leader only) =======================
// One row per follower, with its nextIndex and matchIndex side by side.
// Opened from the leader card. It updates live while it is open.

let openIndexId = null;      // NodeId of the open panel, or null
let indexCells = new Map();  // follower id -> { next, match } table cells

function setCellText(cell, value) {
  const text = formatValue(value);
  if (cell.textContent !== text) cell.textContent = text; // only touch the DOM when it changed
}

function renderIndexPanel(next, match) {
  const body = netById('indexpanel-body');
  const ids = [...new Set([...Object.keys(next), ...Object.keys(match)])].sort(byNodeName);

  // Rebuild the rows only when the set of followers changes. Otherwise update cell by cell.
  if (ids.join('|') !== [...indexCells.keys()].join('|') || ids.length === 0) {
    body.replaceChildren();
    indexCells = new Map();
    if (ids.length === 0) {
      const td = body.insertRow().insertCell();
      td.colSpan = 3;
      td.className = 'muted';
      td.textContent = 'no followers yet';
      return;
    }
    for (const id of ids) {
      const tr = body.insertRow();
      tr.insertCell().textContent = id;
      indexCells.set(id, { next: tr.insertCell(), match: tr.insertCell() });
    }
  }

  for (const id of ids) {
    setCellText(indexCells.get(id).next, next[id]);
    setCellText(indexCells.get(id).match, match[id]);
  }
}

function openIndexPanel(id) {
  const view = nodeViews.get(id);
  if (!view) return;
  closeLogPanel(); // only one floating panel at a time
  openIndexId = id;
  indexCells = new Map();
  netById('indexpanel-title').textContent = `${id} · next / match index`;
  netById('indexpanel').hidden = false;
  renderIndexPanel(view.nextIndex, view.matchIndex);
  netById('indexpanel-close').focus();
}

function closeIndexPanel() {
  openIndexId = null;
  netById('indexpanel').hidden = true;
}







// ======================= NETWORK (messages) =======================

// Go iota values of newraft.MessageType
const MessageType = Object.freeze({
  NewEntry: 0,
  AppendEntries: 1,
  AppendEntriesResponse: 2,
  RequestVote: 3,
  RequestVoteResponse: 4,
  ElectionTimeout: 5,
  HeartbeatTimeout: 6,
  SendAppendEntriesTimeout: 7,
});

// Only these types are shown in the UI. The value is the label (and the legend text).
// Colors live in the CSS: [data-type="<number>"] { --msg: ... }
const NETWORK_TYPES = Object.freeze({
  [MessageType.NewEntry]: 'NewEntry',
  [MessageType.AppendEntries]: 'AppendEntries',
  [MessageType.AppendEntriesResponse]: 'AppendEntriesResponse',
  [MessageType.RequestVote]: 'RequestVote',
  [MessageType.RequestVoteResponse]: 'RequestVoteResponse',
});

// The CSS variable that holds the color of each type (used for the receive effect).
const NETWORK_COLOR_VARS = Object.freeze({
  [MessageType.NewEntry]: '--msg-new-entry',
  [MessageType.AppendEntries]: '--msg-append',
  [MessageType.AppendEntriesResponse]: '--msg-append-response',
  [MessageType.RequestVote]: '--msg-vote',
  [MessageType.RequestVoteResponse]: '--msg-vote-response',
});

// Messages in flight, sorted by DeliveryTick (soonest first, ties by id).
// Each item: { id, tick, sent, type, from, to, row, dot, packet, progress }.
// The id is kept but never shown.
const networkQueue = [];

const NEW_ANIMATION_MS = 1200;

function netById(id) { return document.getElementById(id); }

function netEl(tag, cls, text) {
  const e = document.createElement(tag);
  if (cls) e.className = cls;
  if (text !== undefined) e.textContent = text;
  return e;
}

function typeColor(type) {
  const value = getComputedStyle(document.documentElement).getPropertyValue(NETWORK_COLOR_VARS[type]);
  return value.trim() || '#8093ff';
}

// Reads a SimMessage. If newraft.Message uses other field names, change them here.
function readMessage(sim) {
  const msg = sim.Message ?? {};
  return {
    id: msg.MessageId ?? sim.Id, // SimMessage.Id arrives as 0, MessageId is unique
    tick: sim.DeliveryTick,
    type: msg.Type ?? msg.MessageType,
    from: msg.SenderId ?? msg.From ?? msg.Sender ?? '',
    to: msg.ReceiverId ?? msg.To ?? msg.Receiver ?? '',
  };
}

function handleNewMessage(payload) {
  const list = payload.Messages ?? (Array.isArray(payload) ? payload : []);

  for (const sim of list) {
    const m = readMessage(sim);

    switch (m.type) {
      case MessageType.NewEntry:
      case MessageType.AppendEntries:
      case MessageType.AppendEntriesResponse:
      case MessageType.RequestVote:
      case MessageType.RequestVoteResponse:
        enqueueMessage(m);
        break;

      case MessageType.ElectionTimeout:
        console.log('MsgElectionTimeout arrived:', sim);
        break;
      case MessageType.HeartbeatTimeout:
        console.log('MsgHeartbeatTimeout arrived:', sim);
        break;
      case MessageType.SendAppendEntriesTimeout:
        console.log('MsgSendAppendEntriesTimeout arrived:', sim);
        break;

      default:
        console.warn('Unknown message type:', m.type, sim);
    }
  }
}

function createRow(m) {
  const li = netEl('li');
  li.dataset.type = m.type;
  li.dataset.id = m.id; // kept for the delivery animation, not shown

  const type = netEl('div', 'type');
  type.append(netEl('i', 'dot'), NETWORK_TYPES[m.type]);

  const route = netEl('div', 'route');
  route.append(netEl('span', '', `${m.from} → ${m.to}`), netEl('span', 'tick', `tick ${m.tick}`));

  li.append(type, route);
  return li;
}

function createDot(m) {
  const dot = netEl('i', 'dot');
  dot.dataset.type = m.type;
  dot.dataset.id = m.id;
  return dot;
}

function insertAt(parent, el, index) {
  parent.insertBefore(el, parent.children[index] || null);
}

function enqueueMessage(m) {
  if (networkQueue.some((q) => q.id === m.id)) return; // already in the queue

  const item = { ...m, sent: currentTick, progress: 0, row: createRow(m), dot: createDot(m), packet: null };

  // keep the queue ordered by DeliveryTick, then by id
  const at = networkQueue.findIndex((q) => q.tick > m.tick || (q.tick === m.tick && q.id > m.id));
  const index = at === -1 ? networkQueue.length : at;
  networkQueue.splice(index, 0, item);

  insertAt(netById('msgs'), item.row, index); // explicit view
  insertAt(netById('flow'), item.dot, index); // hidden view

  // entry animation, removed afterwards so it does not replay when switching view
  item.row.classList.add('is-new');
  item.dot.classList.add('is-new');
  setTimeout(() => {
    item.row.classList.remove('is-new');
    item.dot.classList.remove('is-new');
  }, NEW_ANIMATION_MS);

  item.packet = createPacket(item);                       // a dot on the sender...
  if (item.packet) movePacket(item, currentTick + 1);     // ...that starts travelling right away

  updateNetworkMeta();
}

function updateNetworkMeta() {
  netById('net-count').textContent = networkQueue.length;
  netById('net-empty').hidden = networkQueue.length > 0;
}


// ------------ TIME: the current tick and how long a tick takes in real time ------------

let currentTick = 0;
let tickMs = 400;      // estimated real time between two ticks (smoothed)
let lastTickAt = 0;

function onTick(tick) {
  const now = performance.now();
  if (lastTickAt) {
    const elapsed = now - lastTickAt;
    tickMs = Math.min(3000, Math.max(16, tickMs * 0.6 + elapsed * 0.4));
  }
  lastTickAt = now;
  currentTick = tick;
}


// ------------ PACKETS: a message travelling along the line, tick by tick ------------
// The dot leaves the sender at the tick it was sent and reaches the receiver exactly at its
// DeliveryTick. Between ticks it keeps gliding (a CSS transition as long as one tick takes),
// so it moves steadily instead of jumping.

// Center of a node's circle in px, relative to #nodes (null if there is no such node).
function nodeCenter(id) {
  const card = [...netById('nodes').children].find((c) => c.dataset.id === id);
  if (!card) return null;
  const nodes = netById('nodes');
  return {
    x: (parseFloat(card.style.left) / 100) * nodes.clientWidth,
    y: (parseFloat(card.style.top) / 100) * nodes.clientHeight,
  };
}

function ballRadius() {
  const ball = document.querySelector('.node__circle');
  return ball ? ball.offsetWidth / 2 : 40;
}

// For a sender or receiver that is not a node (a client sending a NewEntry): the left edge
// of the router, at the height of the node on the other side.
function networkEdge(y) {
  const nodes = netById('nodes').getBoundingClientRect();
  const net = document.querySelector('.net').getBoundingClientRect();
  return { x: net.left - nodes.left, y };
}

// Straight path from the edge of the sender circle to the edge of the receiver circle.
function packetPath(item) {
  const a = nodeCenter(item.from);
  const b = nodeCenter(item.to);
  if (!a && !b) return null;

  const start = a ?? networkEdge(b.y);
  const end = b ?? networkEdge(a.y);
  const dx = end.x - start.x;
  const dy = end.y - start.y;
  const len = Math.hypot(dx, dy) || 1;
  const gap = ballRadius() + 4;
  const g0 = a ? gap : 0;
  const g1 = b ? gap : 0;

  return {
    x0: start.x + (dx / len) * g0, y0: start.y + (dy / len) * g0,
    x1: end.x - (dx / len) * g1,   y1: end.y - (dy / len) * g1,
  };
}

// 0 at the tick it was sent, 1 at its DeliveryTick.
function progressAt(item, tick) {
  const span = item.tick - item.sent;
  if (span <= 0) return 1;
  return Math.min(1, Math.max(0, (tick - item.sent) / span));
}

function createPacket(item) {
  const path = packetPath(item);
  if (!path) return null;

  const el = netEl('i', 'packet');
  el.dataset.type = item.type;
  netById('packets').append(el);

  el.style.transitionDuration = '0ms';
  el.style.transform = `translate(${path.x0}px, ${path.y0}px)`; // starts on the sender
  void el.offsetWidth; // apply that position before animating
  return el;
}

// Moves the dot to where it must be at `tick`, taking `ms` to get there.
function movePacket(item, tick, ms = tickMs) {
  const path = packetPath(item);
  if (!item.packet || !path) return;

  const p = progressAt(item, tick);
  item.progress = p;
  item.packet.style.transitionDuration = `${ms}ms`;
  item.packet.style.transform =
    `translate(${path.x0 + (path.x1 - path.x0) * p}px, ${path.y0 + (path.y1 - path.y0) * p}px)`;
}

// On every tick, each dot starts gliding toward its position at the NEXT tick, so it
// arrives at the receiver exactly when the delivery tick arrives.
function movePackets(tick) {
  for (const item of networkQueue) movePacket(item, tick + 1);
}

// Layout changed (window resized): put the dots where they belong, without animation.
window.addEventListener('resize', () => {
  for (const item of networkQueue) movePacket(item, currentTick + 1, 0);
});


// ------------ DELIVERY ------------

// Called on every tick: the queue is sorted, so we only look at the front.
function deliverDueMessages(tick) {
  while (networkQueue.length > 0 && networkQueue[0].tick <= tick) {
    deliverMessage(networkQueue[0]);
  }
}

// Delivers a message by id (called from the MessageDelivered event). It does nothing if the
// message was already delivered by deliverDueMessages(), so the effect never plays twice.
function removeMessageById(id) {
  const item = networkQueue.find((q) => String(q.id) === String(id));
  if (item) deliverMessage(item);
}

function deliverMessage(item) {
  const index = networkQueue.indexOf(item);
  if (index === -1) return;

  item.row.remove();
  item.dot.remove();
  networkQueue.splice(index, 1);
  updateNetworkMeta();

  arrive(item);
}

// The dot reaches the receiver: it disappears and the receiver bounces.
function arrive(item) {
  const el = item.packet;
  const ball = nodeBall(item.to);

  const done = () => {
    if (el) el.remove();
    if (ball && !matchMedia('(prefers-reduced-motion: reduce)').matches) {
      receiveBounce(ball, typeColor(item.type));
    }
  };

  if (el && item.progress < 1) {   // delivered before it got there: finish the trip quickly
    movePacket(item, item.tick, 120);
    setTimeout(done, 130);
  } else {
    done();
  }
}

function nodeBall(id) {
  const card = [...netById('nodes').children].find((c) => c.dataset.id === id);
  return card ? card.querySelector('.node__circle') : null;
}

// A small jump, like "I received something".
function receiveBounce(ball, color) {
  ball.animate([
    { transform: 'translateY(0) scale(1)', boxShadow: `0 0 0 0 ${color}` },
    { transform: 'translateY(-14px) scale(1.1)', offset: 0.3 },
    { transform: 'translateY(0) scale(1)', offset: 0.55 },
    { transform: 'translateY(-5px) scale(1.03)', offset: 0.75 },
    { transform: 'translateY(0) scale(1)', boxShadow: '0 0 0 14px transparent' },
  ], { duration: 520, easing: 'ease-out' });
}


// ------------ Lines between nodes + UI setup ------------

// Soft lines from every node to every other node.
// Called at the end of layoutNodes(), after the positions are set.
function drawNodeLinks() {
  const svg = netById('links');
  const cards = [...netById('nodes').children];
  svg.replaceChildren();

  for (let i = 0; i < cards.length; i++) {
    for (let j = i + 1; j < cards.length; j++) {
      const line = document.createElementNS('http://www.w3.org/2000/svg', 'line');
      line.setAttribute('x1', cards[i].style.left);
      line.setAttribute('y1', cards[i].style.top);
      line.setAttribute('x2', cards[j].style.left);
      line.setAttribute('y2', cards[j].style.top);
      svg.append(line);
    }
  }
}

function initNetworkUI() {
  // color legend
  const legend = netById('legend');
  for (const [type, name] of Object.entries(NETWORK_TYPES)) {
    const item = netEl('span');
    item.dataset.type = type;
    item.append(netEl('i', 'dot'), name);
    legend.append(item);
  }

  // Explicit / Hidden switch
  const buttons = document.querySelectorAll('#net-mode button');
  buttons.forEach((b) => b.addEventListener('click', () => {
    document.body.dataset.netMode = b.dataset.mode;
    buttons.forEach((x) => x.setAttribute('aria-pressed', x === b));
  }));

  // floating panels
  netById('fulllog-close').addEventListener('click', closeLogPanel);
  netById('indexpanel-close').addEventListener('click', closeIndexPanel);
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') { closeLogPanel(); closeIndexPanel(); }
  });
}

initNetworkUI();
