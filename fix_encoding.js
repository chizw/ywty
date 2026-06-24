const fs = require('fs');
const path = require('path');

const baseDir = path.join(__dirname, 'web-nuxt', 'pages', 'admin');
const FFFD = '\ufffd';

// For remaining corruptions, we need different search patterns
// because the raw byte corruption (XX YY 3F) loses the following byte
const fixMap = {
  'tickets.vue': [
    // statusMap: the closing ' was lost, so search without it
    ["'待处" + FFFD + "?,", "'待处理',"],
    ["'处理" + FFFD + "?,", "'处理中',"],
    ["'已解" + FFFD + "?,", "'已解决',"],
    ["'已关" + FFFD + "?,", "'已关闭',"],
    // levelMap: entire char replaced
    ["low: '" + FFFD + "?,", "low: '低',"],
    ["medium: '" + FFFD + "?,", "medium: '中',"],
    ["high: '" + FFFD + "?,", "high: '高',"],
    ["urgent: '紧" + FFFD + "?,", "urgent: '紧急',"],
  ],

  'drivers.vue': [
    // providerLabel: closing ' was lost
    ["'七牛" + FFFD + "?,", "'七牛云',"],
    ["'又拍" + FFFD + "?,", "'又拍云',"],
    ["'阿里云邮件推" + FFFD + "?,", "'阿里云邮件推送',"],
    ["'阿里云短" + FFFD + "?,", "'阿里云短信',"],
    ["'腾讯云短" + FFFD + "?,", "'腾讯云短信',"],
    ["'阿里云内容安" + FFFD + "?,", "'阿里云内容安全',"],
    ["'空操" + FFFD + "?,", "'空操作',"],
    // comments and template: closing char lost
    ["短信测试发" + FFFD + "\n", "短信测试发送\n"],
    ["邮件测试发" + FFFD + "\n", "邮件测试发送\n"],
    ["发送失" + FFFD + "')", "发送失败')"],
    ["发送请求已提交（请查看日志/上游" + FFFD + ")", "发送请求已提交（请查看日志/上游）"],
    ["邮件发送请求已提交" + FFFD + "", "邮件发送请求已提交"],  // may not exist
    ["短信发送请求已提交" + FFFD + "", "短信发送请求已提交"],
    // Template text with lost closing chars
    ["存储策略</NuxtLink>。", "存储策略</NuxtLink>。"],  // keep as-is if correct
    // Empty spans - the ? was the only content
    ['text-muted-foreground">' + FFFD + '</span>', 'text-muted-foreground">无</span>'],
    // SMS provider line
    ["SMS provider" + FFFD + "/p>", "SMS provider）</p>"],
    // 邮件发送 line
    ["触发一次发" + FFFD + "/p>", "触发一次发送</p>"],
    // 发送 button
    ['@click="testSMS">' + FFFD + "/Button>", '@click="testSMS">发送</Button>'],
    ['@click="testMail">' + FFFD + "/Button>", '@click="testMail">发送</Button>'],
    // 这是一封来云雾图驿
    ["一封来" + FFFD + "云雾图驿", "一封来自云雾图驿"],
    ["测试邮件" + FFFD + "", "测试邮件。"],  // need to check exact context
    // placeholder 手机号 收件人邮箱
    ['placeholder="' + FFFD + '手机号" class', 'placeholder="手机号" class'],  // may not work
    ['placeholder="' + FFFD + '收件人邮箱" class', 'placeholder="收件人邮箱" class'],
    ['placeholder="手机' + FFFD + ' class', 'placeholder="手机号" class'],
    ['placeholder="收件人邮' + FFFD + ' class', 'placeholder="收件人邮箱" class'],
    // 短信测试标题
    ['已配置 SMS provider' + FFFD + '/p>', '已配置 SMS provider）</p>'],
  ],

  'storage.vue': [
    // comment
    ["增删改查" + FFFD + ")", "增删改查）"],
    // providerLabel
    ["'七牛" + FFFD + "?,", "'七牛云',"],
    ["'又拍" + FFFD + "?,", "'又拍云',"],
    ["'阿里" + FFFD + "?OSS", "'阿里云OSS"],
    ["'腾讯" + FFFD + "?COS", "'腾讯云COS"],
    // messages
    ["'已更" + FFFD + "?,", "'已更新',"],
    ["'已创" + FFFD + "?,", "'已创建',"],
    ["'已删" + FFFD + "?,", "'已删除',"],
    ["请填写名" + FFFD + "?)", "请填写名称')"],
    ["options 必须是合" + FFFD + "?JSON", "options 必须是合法JSON"],
    // Template
    ["按策略 ID 调用" + FFFD + "   </p>", "按策略 ID 调用。    </p>"],
    ["\u201c新建策略\u201d开始" + FFFD + "   </div>", "\u201c新建策略\u201d开始。    </div>"],
    ["如：阿里" + FFFD + "?OSS 主站", "如：阿里云 OSS 主站"],
    ["（{{ d }}" + FFFD + "/SelectItem>", "（{{ d }}）</SelectItem>"],
    ["Options（JSON" + FFFD + "/Label>", "Options（JSON）</Label>"],
    ["不同驱动" + FFFD + "?options", "不同驱动的 options"],
    ["参考驱动文档填写" + FFFD + "/p>", "参考驱动文档填写。</p>"],
  ],

  'pages.vue': [
    // comment
    ["单页管" + FFFD + "\ndefinePageMeta", "单页管理\ndefinePageMeta"],
    // placeholder
    ['placeholder="' + FFFD + 'about"', 'placeholder="如 about"'],
  ],

  'photos.vue': [
    // comment
    ["图片管" + FFFD + "\ndefinePageMeta", "图片管理\ndefinePageMeta"],
    // counter
    ["}} " + FFFD + " </span>", "}} 张 </span>"],
  ],

  'reports.vue': [
    // comment  
    ["举报管" + FFFD + "\ndefinePageMeta", "举报管理\ndefinePageMeta"],
    // 无说明
    ["（无说明" + FFFD + ")", "（无说明）"],
  ],

  'groups.vue': [
    // new group button text
    ["新建角色" + FFFD + " }}", "新建角色组 }}"],
    // register default label
    ["注册时默认使" + FFFD + "/Label>", "注册时默认使用</Label>"],
    // delete confirm
    ["用户角色绑定也会被解除" + FFFD + "/p>", "用户角色绑定也会被解除。</p>"],
  ],

  'license.vue': [
    // 免费版
    ["免费" + FFFD + "' }}", "免费版' }}"],
    // 已激活 已过期 未激活
    ["'已激" + FFFD + " : ", "'已激活' : "],
    ["'已过" + FFFD + " : ", "'已过期' : "],
    ["'未激" + FFFD + "'", "'未激活'"],
    // 无限制
    ["无限制" + FFFD + "' }}", "无限制' }}"],  // hmm, check context
    // 最大存储空间
    ["最大存储空" + FFFD + "            </div>", "最大存储空间            </div>"],
    // 已启用功能
    ["已启用功" + FFFD + "          </div>", "已启用功能          </div>"],
    // 激活按钮
    ["? '激" + FFFD + "License' : '激" + FFFD + "License'", "? '激活 License' : '激活 License'"],
    // 请输入 License Key
    ['placeholder="请输' + FFFD + 'License Key"', 'placeholder="请输入 License Key"'],
    // 激活按钮  
    ['@click="activate">' + FFFD + "/Button>", '@click="activate">激活</Button>'],
    // 状态 label
    ["状" + FFFD + "            </div>", "状态            </div>"],
  ],

  'violations.vue': [
    // comment
    ["违规记录管" + FFFD + "\ndefinePageMeta", "违规记录管理\ndefinePageMeta"],
    // statusMap
    ["'待处" + FFFD + "?,", "'待处理',"],
    ["'已处" + FFFD + "?,", "'已处理',"],
    ["'已忽" + FFFD + "?,", "'已忽略',"],
  ],

  'users.vue': [
    // remaining: 共 X 个用户 label
    [FFFD + "{{ meta?.total ?? users.length }} 个用户", "共 {{ meta?.total ?? users.length }} 个用户"],
    // 是/否
    ["{{ u.is_admin ? '" + FFFD + " : '" + FFFD + " }}", "{{ u.is_admin ? '是' : '否' }}"],
  ],

  'index.vue': [
    // remaining: P2 corruptions (extra ? after 报)
    ['举报?          </div>', '举报          </div>'],
  ],

  'feedbacks.vue': [
    ["意见管" + FFFD + "\ndefinePageMeta", "意见管理\ndefinePageMeta"],
    ["反馈管" + FFFD + "\ndefinePageMeta", "反馈管理\ndefinePageMeta"],
  ],

  'notices.vue': [
    // remaining
    ["确定删除该通知" + FFFD + "/p>", "确定删除该通知？</p>"],
  ],
};

let totalFixed = 0;

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
      missed.push(search.substring(0, 30).replace(/\ufffd/g, 'FFFD'));
    }
  }
  
  if (content !== original) {
    fs.writeFileSync(filePath, content, 'utf8');
    console.log(`${file}: ${fileFixCount} applied`);
    if (missed.length > 0) console.log(`  missed: ${missed.join('; ')}`);
    totalFixed += fileFixCount;
  } else {
    console.log(`${file}: no changes (${missed.length} patterns not found)`);
    if (missed.length > 0) console.log(`  missed: ${missed.join('; ')}`);
  }
}

console.log(`\nTotal: ${totalFixed} replacements`);
