const fs = require('fs');
const path = require('path');

const filePath = path.join(__dirname, 'web-nuxt', 'pages', 'admin', 'tickets.vue');
const content = fs.readFileSync(filePath, 'utf8');
const FFFD = '\ufffd';

// Check if the search strings exist
const searches = [
  '待处' + FFFD + "?'",
  '处理' + FFFD + "?'",
  '已解' + FFFD + "?'",
  '已关' + FFFD + "?'",
  FFFD + "?'",  // for levelMap
];

for (const s of searches) {
  const idx = content.indexOf(s);
  if (idx >= 0) {
    const ctx = content.substring(Math.max(0, idx - 10), idx + s.length + 10);
    console.log(`FOUND: "${s.replace(/\ufffd/g, 'FFFD')}" at ${idx}: [${ctx.replace(/\ufffd/g, 'FFFD')}]`);
  } else {
    console.log(`NOT FOUND: "${s.replace(/\ufffd/g, 'FFFD')}"`);
  }
}

// Also check what the statusMap lines look like
const lines = content.split('\n');
for (let i = 11; i <= 24; i++) {
  console.log(`L${i+1}: ${lines[i].replace(/\ufffd/g, 'FFFD')}`);
}
