# 5th-bytedance-youth-camp-tiktok

### 第五届字节跳动青训营抖音项目

- 初始化了数据交互的目录结构以及代码
- 创建了数据库表（包含外键、依赖），参考了 <a href="https://github.com/a76yyyy/tiktok">第三届字节青训营的内容</a>
- <a href="https://www.apifox.cn/apidoc/shared-09d88f32-0b6c-4157-9d07-a36d32d7a75c">极简版抖音APP接口文档</a>

| 数据库表名                    | Golang 类名                                | 备注           |
|--------------------------|------------------------------------------|--------------|
| comment                  | Comment                                  | 评论           |
| messages                 | Message                                  | 聊天消息         |
| relation                 | FollowRelation (/relation.go)            | 社交（粉丝、关注、朋友） |
| user                     | User                                     | 用户           |
| video                    | Video (/feed.go)                         | 视频           |
| user_favorite_videos     | FavoriteVideoRelation (/favorite.go)     | 视频点赞记录       |
| _user_favorite_comments_ | _FavoriteCommentRelation (/favorite.go)_ | _评论点赞记录_     |

注：

- 聊天消息为本届项目新增内容，评论点赞记录未开发。
- 朋友为互相关注的用户

#### 启动/停止运行

- 启动服务：`sh startup.sh`
- 停止运行：`sh shutdown.sh`
