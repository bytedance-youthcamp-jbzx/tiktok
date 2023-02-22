# bytedance-youthcamp-tiktok

第五届字节跳动青训营抖音项目

- 初始化了数据交互的目录结构以及代码
- 创建了数据库表（包含外键、依赖），参考了 <a href="https://github.com/a76yyyy/tiktok">第三届字节青训营的内容</a>
- <a href="https://www.apifox.cn/apidoc/shared-09d88f32-0b6c-4157-9d07-a36d32d7a75c">极简版抖音APP接口文档</a>

| 数据库表名    | Golang 类名                       | 备注     |
|----------|---------------------------------|--------|
| comment  | Comment                         | 评论     |
| messages | Message                         | *聊天消息  |
| relation | FollowRelation (/relation.go)   | 关注     |
| user     | User                            | 用户     |
| video    | Video (/feed.go)                | 视频     |
| favorite | FavoriteRelation (/favorite.go) | 用户点赞记录 |

*聊天消息为本届项目新增内容

#### 启动/停止运行

- 启动服务：`sh startup.sh`
- 停止运行：`sh shutdown.sh`
