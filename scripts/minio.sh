docker run -p 9000:9000 -p 9090:9090   \
    --net=bridge      \
    --name minio     \
    -d --restart=always    \
    -e "MINIO_ACCESS_KEY=tiktokMinio"    \
    -e "MINIO_SECRET_KEY=tiktokMinio"    \
    -v /home/minio/data:/data   \
    -v /home/minio/config:/root/.minio  \
    minio/minio server    \
    /data --console-address ":9090" -address ":9000"
