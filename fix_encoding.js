const fs = require('fs');
const path = require('path');

const baseDir = 'd:\\projects\\ywty\\web-nuxt\\pages\\admin';

// Replacement map: each entry is [file, searchContext, replaceContext]
// We find the corruption by looking for the ? character after Chinese text,
// and use surrounding context to determine what to replace.

const replacements = [
  // index.vue
  ['index.vue', '仪表\ufffd?/h1>', '仪表盘</h1>'],
  ['index.vue', '处理举报?', '处理举报'],
  ['index.vue', '总举报?', '总举报'],
  
  // users.vue
  ['users.vue', '用户管理\ndefinePageMeta', '用户管理\ndefinePageMeta'],
  ['users.vue', '已保存\n', '已保存\n'],
  ['users.vue', '个用户</span>', '个用户</span>'],
  ['users.vue', '搜索用户名/邮箱/姓名', '搜索用户名/邮箱/姓名'],
  ['users.vue', '用户名</TableHead>', '用户名</TableHead>'],
  ['users.vue', '状态</TableHead>', '状态</TableHead>'],
  ['users.vue', "管理员</TableHead>", "管理员</TableHead>"],
  ["users.vue", "'是' : '否'", "'是' : '否'"],
  ['users.vue', '上一页</Button>', '上一页</Button>'],
  ['users.vue', '第 {{ meta.current_page }} / {{ meta.last_page }} 页</span>', '第 {{ meta.current_page }} / {{ meta.last_page }} 页</span>'],
  ['users.vue', '下一页</Button>', '下一页</Button>'],
  ['users.vue', '状态</Label>', '状态</Label>'],
  ['users.vue', '设为管理员</Label>', '设为管理员</Label>'],
  ['users.vue', '角色组</Label>', '角色组</Label>'],
  ['users.vue', "不修改 />", "不修改 />"],
  ['users.vue', '不修改</SelectItem>', '不修改</SelectItem>'],

  // photos.vue
  ['photos.vue', '图片管理', '图片管理'],
  ['photos.vue', '搜索文件名/原始名', '搜索文件名/原始名'],
  ['photos.vue', '张</span>', '张</span>'],
  ['photos.vue', '上一页</Button>', '上一页</Button>'],
  ['photos.vue', '第 {{ meta.current_page }} / {{ meta.last_page }} 页</span>', '第 {{ meta.current_page }} / {{ meta.last_page }} 页</span>'],
  ['photos.vue', '下一页</Button>', '下一页</Button>'],
  ['photos.vue', '不可恢复。', '不可恢复。'],

  // tickets.vue
  ['tickets.vue', '工单管理', '工单管理'],
  ['tickets.vue', '待处理', '待处理'],
  ['tickets.vue', '处理中', '处理中'],
  ['tickets.vue', '已解决', '已解决'],
  ['tickets.vue', '已关闭', '已关闭'],
  ['tickets.vue', '低', '低'],
  ['tickets.vue', '中', '中'],
  ['tickets.vue', '高', '高'],
  ['tickets.vue', '紧急', '紧急'],
  ['tickets.vue', '优先级</TableHead>', '优先级</TableHead>'],
  ['tickets.vue', '状态</TableHead>', '状态</TableHead>'],
];

// For each file, read as bytes, find corruption patterns, and fix
const files = fs.readdirSync(baseDir).filter(f => f.endsWith('.vue'));

for (const file of files) {
  const filePath = path.join(baseDir, file);
  let bytes = fs.readFileSync(filePath);
  let modified = false;
  
  // Strategy: find all occurrences of 0xEF 0xBF 0xBD 0x3F and replace with appropriate char
  // Also find standalone 0x3F that comes after valid Chinese chars (the extra ? pattern)
  
  // For the EF BF BD 3F pattern, we need to know what character it should be
  // We'll collect all positions and then use context to determine replacements
  
  const results = [];
  let i = 0;
  while (i < bytes.length - 3) {
    // Pattern 1: EF BF BD 3F
    if (bytes[i] === 0xEF && bytes[i+1] === 0xBF && bytes[i+2] === 0xBD && bytes[i+3] === 0x3F) {
      // Get context before
      const beforeStart = Math.max(0, i - 30);
      const before = bytes.slice(beforeStart, i).toString('utf8');
      const afterStart = Math.min(bytes.length, i + 4);
      const afterEnd = Math.min(bytes.length, i + 10);
      const after = bytes.slice(afterStart, afterEnd).toString('utf8');
      results.push({ pos: i, type: 'FFFD', before, after });
    }
    i++;
  }
  
  // Now scan for standalone ? (0x3F) that appear after Chinese characters
  // A Chinese character in UTF-8 has bytes in range E0-EF xx xx
  // Look for: [valid 3-byte Chinese char] followed by 0x3F
  i = 0;
  while (i < bytes.length - 3) {
    if (bytes[i] >= 0xE0 && bytes[i] <= 0xEF && 
        bytes[i+1] >= 0x80 && bytes[i+1] <= 0xBF &&
        bytes[i+2] >= 0x80 && bytes[i+2] <= 0xBF) {
      // Valid 3-byte Chinese char at i..i+2
      if (i + 3 < bytes.length && bytes[i+3] === 0x3F) {
        const charBuf = Buffer.from([bytes[i], bytes[i+1], bytes[i+2]]);
        const char = charBuf.toString('utf8');
        const beforeStart = Math.max(0, i - 30);
        const before = bytes.slice(beforeStart, i).toString('utf8');
        const afterStart = Math.min(bytes.length, i + 4);
        const afterEnd = Math.min(bytes.length, i + 10);
        const after = bytes.slice(afterStart, afterEnd).toString('utf8');
        
        // Only count if this ? isn't part of valid code (like ?? in JS)
        // Check if the next byte after ? is also a valid continuation
        const isCodeContext = (after.startsWith("'") || after.startsWith('"') || after.startsWith('}') || after.startsWith('('));
        
        // We need to check if this ? is part of a ?? operator or ternary or similar
        // Chinese chars are typically followed by Chinese chars, not ?, in templates
        results.push({ pos: i, type: 'EXTRA_?', char, before, after, nextByte: i + 4 < bytes.length ? bytes[i+4] : null });
      }
    }
    i++;
  }
  
  if (results.length > 0) {
    console.log(`\n${file}: Found ${results.length} potential corruptions`);
    for (const r of results) {
      console.log(`  ${r.type} at byte ${r.pos}: char=${r.char || 'FFFD'} before=[${r.before.slice(-15)}] after=[${r.after}]`);
    }
  }
}
