const fs = require('fs');
const path = require('path');

const baseDir = path.join(__dirname, 'web-nuxt', 'pages', 'admin');
const R = '\ufffd'; // U+FFFD replacement character

// All corruption follows the same pattern: \ufffd? in the file
// (corrupted last byte of multi-byte UTF-8 char becomes U+FFFD + '?')
// Search strings include the ? that follows \ufffd in every case.

const fixMap = {
  'pages.vue': [
    // comment: 单页管理
    ['单页管' + R + '?definePageMeta', '单页管理\ndefinePageMeta'],
    // placeholder: 如 about
    ['placeholder="' + R + '?about"', 'placeholder="如 about"'],
  ],

  'photos.vue': [
    // comment: 图片管理
    ['图片管' + R + '?definePageMeta', '图片管理\ndefinePageMeta'],
    // counter: 张
    ['}} ' + R + '?/span>', '}} 张</span>'],
  ],

  'reports.vue': [
    // comment: 举报管理
    ['举报管' + R + '?definePageMeta', '举报管理\ndefinePageMeta'],
    // 无说明
    ['（无说明' + R + '? }}', '（无说明） }}'],
  ],

  'feedbacks.vue': [
    // comment: may be 意见管理 or 反馈管理
    ['意见反馈管' + R + '?', '意见反馈管理'],
    ['反馈管' + R + '?definePageMeta', '反馈管理\ndefinePageMeta'],
  ],

  'tickets.vue': [
    // comment: 工单管理
    ['工单管' + R + '?', '工单管理'],
  ],

  'violations.vue': [
    // comment: 违规记录管理
    ['违规记录管' + R + '?', '违规记录管理'],
  ],

  'users.vue': [
    // 共 X 个用户
    [R + '?{{ meta?.total ?? users.length }} 个用户', '共 {{ meta?.total ?? users.length }} 个用户'],
    // 是/否
    ["{{ u.is_admin ? '" + R + "? : '" + R + "? }}", "{{ u.is_admin ? '是' : '否' }}"],
  ],

  'index.vue': [
    // 多余的 ?
    ['举报?          </div>', '举报          </div>'],
  ],

  'notices.vue': [
    // ？
    ['确定删除该通知' + R + '?/p>', '确定删除该通知？</p>'],
  ],

  'storage.vue': [
    // comment: 增删改查）
    ['增删改查' + R + '?', '增删改查）'],
    // messages
    ["'已更" + R + "?)", "'已更新')"],
    ["'已创" + R + "?)", "'已创建')"],
    ["'已删" + R + "?)", "'已删除')"],
    // 请填写名称
    ['请填写名' + R + "?)", "请填写名称')"],
    // options 必须是合法JSON
    ['options 必须是合' + R + '?JSON', 'options 必须是合法JSON'],
    // template: 调用。
    ['按策略 ID 调用' + R + '?    </p>', '按策略 ID 调用。    </p>'],
    // 开始。
    ['"新建策略"开始' + R + '?    </div>', '"新建策略"开始。    </div>'],
    // SelectItem: ）
    ['（{{ d }}' + R + '?/SelectItem>', '（{{ d }}）</SelectItem>'],
    // Label: ）
    ['Options（JSON' + R + '?/Label>', 'Options（JSON）</Label>'],
    // 文档填写。
    ['参考驱动文档填写' + R + '?/p>', '参考驱动文档填写。</p>'],
  ],

  'drivers.vue': [
    // comment: 短信测试发送
    ['短信测试发' + R + '?const sms', '短信测试发送\nconst sms'],
    // 验证码：
    ['您的验证码' + R + '?123456', '您的验证码：123456'],
    // 发送失败
    ["'发送失" + R + "?)", "'发送失败')"],
    // comment: 邮件测试发送
    ['邮件测试发' + R + '?const mai', '邮件测试发送\nconst mai'],
    // 一封来自
    ['一封来' + R + '?云雾图驿', '一封来自云雾图驿'],
    // 测试邮件。 (literal 。? where ? is extra)
    ['测试邮件。? }', '测试邮件。 }'],
    // 测试邮件。  (if \ufffd? variant exists)
    ['测试邮件' + R + '? }', '测试邮件。 }'],
    // </NuxtLink>。
    ['</NuxtLink>' + R + '?      驱动实际', '</NuxtLink>。      驱动实际'],
    // ）。
    ['）' + R + '?    </p>', '）    </p>'],
    // 无</span>
    ['">' + R + '?/span>', '">无</span>'],
    // SMS provider）
    ['SMS provider' + R + '?/p>', 'SMS provider）</p>'],
    // 触发一次发送
    ['触发一次发' + R + '?/p>', '触发一次发送</p>'],
    // 发送</Button>
    ['testMail">发' + R + '?/Button>', 'testMail">发送</Button>'],
  ],

  'groups.vue': [
    // 新建角色组
    ['新建角色' + R + '? }}', '新建角色组 }}'],
    // 注册时默认使用
    ['注册时默认使' + R + '?/Label>', '注册时默认使用</Label>'],
    // 。
    ['用户角色绑定也会被解除' + R + '?/p>', '用户角色绑定也会被解除。</p>'],
  ],

  'license.vue': [
    // 免费版
    ["免费" + R + "? }}", "免费版' }}"],
    // 已激活
    ["'已激" + R + "? : ", "'已激活' : "],
    // 已过期
    ["'已过" + R + "? : ", "'已过期' : "],
    // 未激活
    ["'未激" + R + "? }}", "'未激活' }}"],
    // 无限制
    ["无限" + R + "? }}", "无限制' }}"],
    // 请输入
    ['请输' + R + '?License Key', '请输入 License Key'],
  ],
};

let totalFixed = 0;
const results = {};

for (const [file, fixes] of Object.entries(fixMap)) {
  const filePath = path.join(baseDir, file);
  if (!fs.existsSync(filePath)) {
    console.log(`SKIP ${file}: not found`);
    continue;
  }

  let content = fs.readFileSync(filePath, 'utf8');
  const original = content;
  let fileFixCount = 0;
  let missed = [];

  for (const [search, replace] of fixes) {
    if (content.includes(search)) {
      content = content.split(search).join(replace);
      fileFixCount++;
    } else {
      // Show truncated search for debugging
      const s = search.substring(0, 40).replace(/\ufffd/g, '[FFFD]');
      missed.push(s);
    }
  }

  if (content !== original) {
    fs.writeFileSync(filePath, content, 'utf8');
    console.log(`${file}: ${fileFixCount} fixed`);
    if (missed.length > 0) console.log(`  missed: ${missed.join(' | ')}`);
    totalFixed += fileFixCount;
  } else {
    if (missed.length > 0) {
      console.log(`${file}: 0 fixed, ${missed.length} missed`);
      console.log(`  missed: ${missed.join(' | ')}`);
    } else {
      console.log(`${file}: already clean`);
    }
  }
}

console.log(`\nTotal: ${totalFixed} replacements`);

// Verification: scan for remaining \ufffd
console.log('\n--- Verification: remaining \\ufffd ---');
let remaining = 0;
for (const file of fs.readdirSync(baseDir).filter(f => f.endsWith('.vue'))) {
  const c = fs.readFileSync(path.join(baseDir, file), 'utf8');
  const count = (c.match(/\ufffd/g) || []).length;
  if (count > 0) {
    console.log(`  ${file}: ${count} remaining`);
    remaining += count;
  }
}
if (remaining === 0) console.log('  All clean!');
else console.log(`  Total remaining: ${remaining}`);
