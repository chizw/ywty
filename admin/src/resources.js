import { h } from 'vue';
import { NTag } from 'naive-ui';
import { put, post, del } from './api';

const tag = (v, ok, bad) =>
  h(NTag, { size: 'small', type: v === ok ? 'success' : v === bad ? 'error' : 'warning' }, () => v);

const dt = (r) => (r.created_at || '').replace('T', ' ').slice(0, 19);

export const resources = {
  users: {
    title: '用户',
    description: '管理注册用户：冻结账号、重置密码、删除',
    listURL: '/users',
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'username', label: '用户名' },
      { key: 'name', label: '昵称' },
      { key: 'email', label: '邮箱' },
      { key: 'photo_count', label: '图片数', width: '80px' },
      { key: 'is_admin', label: '管理员', width: '80px', render: (r) => (r.is_admin ? '是' : '否') },
      { key: 'status', label: '状态', width: '90px', render: (r) => tag(r.status, 'normal', 'frozen') },
      { key: 'created_at', label: '注册时间', width: '170px', render: dt },
    ],
    actions: [
      {
        label: '冻结/解冻', type: 'warning', confirm: '切换该用户的冻结状态？',
        onClick: (row, reload) => put('/users/' + row.id, { status: row.status === 'normal' ? 'frozen' : 'normal' }).then(reload),
      },
      {
        label: '重置密码',
        confirm: '将该用户密码重置为随机密码？',
        onClick: (row, reload) =>
          post('/users/' + row.id, { password: 'ywty' + Math.random().toString(36).slice(2, 10) + '!A' }).then(reload),
      },
      {
        label: '删除', type: 'danger', confirm: '确定删除该用户？',
        onClick: (row, reload) => del('/users/' + row.id).then(reload),
      },
    ],
  },

  photos: {
    title: '图片',
    description: '管理全部上传图片：审核状态、删除（同时清理物理文件）',
    listURL: '/photos',
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'name', label: '名称' },
      { key: 'pathname', label: '路径', ellipsis: true },
      { key: 'size', label: '大小(KB)', width: '90px', render: (r) => Number(r.size).toFixed(1) },
      { key: 'dims', label: '尺寸', width: '100px', render: (r) => `${r.width}x${r.height}` },
      { key: 'is_public', label: '公开', width: '70px', render: (r) => (r.is_public ? '是' : '否') },
      { key: 'status', label: '状态', width: '90px', render: (r) => tag(r.status, 'normal', 'violation') },
      { key: 'created_at', label: '时间', width: '170px', render: dt },
    ],
    actions: [
      {
        label: '标记违规', type: 'warning', confirm: '将该图片标记为违规？',
        onClick: (row, reload) => put(`/photos/${row.id}/status`, { status: 'violation' }).then(reload),
      },
      {
        label: '恢复正常', confirm: '恢复正常状态？',
        onClick: (row, reload) => put(`/photos/${row.id}/status`, { status: 'normal' }).then(reload),
      },
      {
        label: '删除', type: 'danger', confirm: '删除会同时清理物理文件，确认？',
        onClick: (row, reload) => del('/photos/' + row.id).then(reload),
      },
    ],
  },

  albums: {
    title: '相册',
    description: '管理用户相册',
    listURL: '/albums',
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'user_id', label: '用户ID', width: '80px' },
      { key: 'name', label: '名称' },
      { key: 'intro', label: '介绍', ellipsis: true },
      { key: 'photo_count', label: '图片数', width: '80px' },
      { key: 'is_public', label: '公开', width: '70px', render: (r) => (r.is_public ? '是' : '否') },
      { key: 'created_at', label: '创建时间', width: '170px', render: dt },
    ],
    actions: [
      {
        label: '删除', type: 'danger', confirm: '确认删除相册？',
        onClick: (row, reload) => del('/albums/' + row.id).then(reload),
      },
    ],
  },

  groups: {
    title: '角色组',
    description: '用户角色组与上传配额',
    listURL: '/groups',
    searchable: false,
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'name', label: '名称' },
      { key: 'intro', label: '描述', ellipsis: true },
      { key: 'is_default', label: '默认组', width: '80px', render: (r) => (r.is_default ? '是' : '否') },
      { key: 'is_guest', label: '游客组', width: '80px', render: (r) => (r.is_guest ? '是' : '否') },
    ],
  },

  shares: {
    title: '分享',
    description: '用户创建的分享链接',
    listURL: '/shares',
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'user_id', label: '用户ID', width: '80px' },
      { key: 'type', label: '类型', width: '80px' },
      { key: 'slug', label: 'Slug' },
      { key: 'view_count', label: '浏览', width: '80px' },
      { key: 'expired_at', label: '过期时间', width: '170px', render: (r) => (r.expired_at || '').replace('T', ' ').slice(0, 19) },
      { key: 'created_at', label: '创建时间', width: '170px', render: dt },
    ],
    actions: [
      {
        label: '删除', type: 'danger', confirm: '确认删除分享？',
        onClick: (row, reload) => del('/shares/' + row.id).then(reload),
      },
    ],
  },

  plans: {
    title: '套餐',
    description: '订阅套餐与价格阶梯',
    listURL: '/plans',
    searchable: false,
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'type', label: '类型', width: '90px' },
      { key: 'name', label: '名称' },
      { key: 'badge', label: '徽章', width: '90px' },
      { key: 'is_up', label: '上架', width: '70px', render: (r) => (r.is_up ? '是' : '否') },
      { key: 'prices', label: '价格(分)', render: (r) => (r.prices || []).map((p) => `${p.name} ${p.price}分/${p.duration}分`).join('、') },
    ],
    actions: [
      {
        label: '下架/上架', confirm: '切换上架状态？',
        onClick: (row, reload) => put('/plans/' + row.id, { is_up: !row.is_up }).then(reload),
      },
      {
        label: '删除', type: 'danger', confirm: '确认删除套餐？',
        onClick: (row, reload) => del('/plans/' + row.id).then(reload),
      },
    ],
  },

  orders: {
    title: '订单',
    description: '支付订单记录',
    listURL: '/orders',
    columns: [
      { key: 'trade_no', label: '订单号', width: '190px' },
      { key: 'amount', label: '金额(分)', width: '90px' },
      { key: 'deduct_amount', label: '抵扣(分)', width: '90px' },
      { key: 'pay_method', label: '支付方式', width: '100px' },
      { key: 'status', label: '状态', width: '100px', render: (r) => tag(r.status, 'paid') },
      { key: 'created_at', label: '时间', width: '170px', render: dt },
    ],
  },

  coupons: {
    title: '优惠码',
    description: '折扣券码管理',
    listURL: '/coupons',
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'name', label: '名称' },
      { key: 'code', label: '券码' },
      { key: 'type', label: '类型', width: '90px' },
      { key: 'value', label: '面值/折扣', width: '100px' },
      { key: 'usage_limit', label: '限用次数', width: '90px' },
      { key: 'expired_at', label: '过期时间', width: '170px', render: (r) => (r.expired_at || '').replace('T', ' ').slice(0, 19) },
    ],
    actions: [
      {
        label: '删除', type: 'danger', confirm: '确认删除优惠码？',
        onClick: (row, reload) => del('/coupons/' + row.id).then(reload),
      },
    ],
  },

  notices: {
    title: '公告',
    description: '全站公告管理',
    listURL: '/notices',
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'title', label: '标题' },
      { key: 'content', label: '内容', ellipsis: true },
      { key: 'sort', label: '排序', width: '70px' },
    ],
    actions: [
      {
        label: '删除', type: 'danger', confirm: '确认删除公告？',
        onClick: (row, reload) => del('/notices/' + row.id).then(reload),
      },
    ],
  },

  pages: {
    title: '页面',
    description: '自定义独立页面',
    listURL: '/pages',
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'name', label: '名称' },
      { key: 'title', label: '标题' },
      { key: 'slug', label: 'Slug' },
      { key: 'type', label: '类型', width: '90px' },
      { key: 'is_show', label: '显示', width: '70px', render: (r) => (r.is_show ? '是' : '否') },
    ],
    actions: [
      {
        label: '删除', type: 'danger', confirm: '确认删除页面？',
        onClick: (row, reload) => del('/pages/' + row.id).then(reload),
      },
    ],
  },

  tickets: {
    title: '工单',
    description: '用户工单',
    listURL: '/tickets',
    columns: [
      { key: 'issue_no', label: '工单号', width: '190px' },
      { key: 'title', label: '标题' },
      { key: 'username', label: '用户', width: '120px' },
      { key: 'level', label: '级别', width: '80px' },
      { key: 'status', label: '状态', width: '110px', render: (r) => tag(r.status, 'completed') },
    ],
  },

  reports: {
    title: '举报',
    description: '内容举报记录',
    listURL: '/reports',
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'reportable_type', label: '类型', width: '180px' },
      { key: 'reportable_id', label: '目标ID', width: '80px' },
      { key: 'content', label: '内容', ellipsis: true },
      { key: 'status', label: '状态', width: '100px', render: (r) => tag(r.status, 'handled') },
    ],
    actions: [
      {
        label: '标记已处理', confirm: '标记为已处理？',
        onClick: (row, reload) => put('/reports/' + row.id, {}).then(reload),
      },
    ],
  },

  violations: {
    title: '违规',
    description: '图片违规记录',
    listURL: '/violations',
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'user_id', label: '用户ID', width: '80px' },
      { key: 'photo_id', label: '图片ID', width: '80px' },
      { key: 'reason', label: '原因', ellipsis: true },
      { key: 'status', label: '状态', width: '100px', render: (r) => tag(r.status, 'handled') },
      { key: 'created_at', label: '时间', width: '170px', render: dt },
    ],
    actions: [
      {
        label: '标记已处理', confirm: '标记为已处理？',
        onClick: (row, reload) => put('/violations/' + row.id, {}).then(reload),
      },
    ],
  },

  feedbacks: {
    title: '反馈',
    description: '意见与 DMCA 反馈',
    listURL: '/feedbacks',
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'type', label: '类型', width: '90px' },
      { key: 'title', label: '标题' },
      { key: 'name', label: '姓名', width: '100px' },
      { key: 'email', label: '邮箱', width: '180px' },
      { key: 'content', label: '内容', ellipsis: true },
      { key: 'created_at', label: '时间', width: '170px', render: dt },
    ],
  },

  storages: {
    title: '储存策略',
    description: '上传文件使用的储存配置',
    listURL: '/storages',
    searchable: false,
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'name', label: '名称' },
      { key: 'prefix', label: '前缀' },
      { key: 'provider', label: '提供者', width: '110px' },
      { key: 'options', label: '配置', ellipsis: true, render: (r) => JSON.stringify(r.options) },
    ],
    actions: [
      {
        label: '删除', type: 'danger', confirm: '确认删除储存策略？',
        onClick: (row, reload) => del('/storages/' + row.id).then(reload),
      },
    ],
  },

  drivers: {
    title: '驱动',
    description: '邮件/支付/短信等驱动配置',
    listURL: '/drivers',
    searchable: false,
    columns: [
      { key: 'id', label: 'ID', width: '70px' },
      { key: 'type', label: '类型', width: '110px' },
      { key: 'name', label: '名称' },
      { key: 'options', label: '配置', ellipsis: true, render: (r) => JSON.stringify(r.options) },
    ],
    actions: [
      {
        label: '删除', type: 'danger', confirm: '确认删除驱动？',
        onClick: (row, reload) => del('/drivers/' + row.id).then(reload),
      },
    ],
  },
};
