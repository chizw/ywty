const fs = require('fs');
const path = require('path');
const baseDir = path.join(__dirname, 'web-nuxt', 'pages', 'admin');
const output = [];

function log(msg) { output.push(msg); }

for (const file of fs.readdirSync(baseDir).filter(f => f.endsWith('.vue'))) {
  const content = fs.readFileSync(path.join(baseDir, file), 'utf8');
  const lines = content.split('\n');
  let found = false;
  
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    // Check for U+FFFD
    if (line.includes('\ufffd')) {
      if (!found) { log(`\n=== ${file} ===`); found = true; }
      log(`  L${i+1}: ${line.trim().substring(0, 100)}`);
    }
    // Check for suspicious ? after Chinese chars (but not in JS code)
    // Look for CJK char + ? in non-code contexts
    const matches = line.match(/[\u4e00-\u9fff]\?[^'")\]}>a-zA-Z0-9]/g);
    if (matches) {
      // Filter out legitimate ? usage
      const isCodeLine = line.includes('?.') || line.includes('??') || line.includes('? ') && line.includes('=');
      if (!isCodeLine) {
        if (!found) { log(`\n=== ${file} ===`); found = true; }
        log(`  L${i+1} (suspect ?): ${line.trim().substring(0, 100)}`);
      }
    }
  }
}

fs.writeFileSync(path.join(__dirname, 'scan2_output.txt'), output.join('\n'), 'utf8');
console.log(`Scan complete. Output: scan2_output.txt`);
