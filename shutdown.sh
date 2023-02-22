session_name=dousheng

tmux has-session -t $session_name
if [ $? -eq 0 ];then
    for i in $(seq 0 6)
    do
        echo "closing window: $i"
        tmux send-keys -t $session_name:$i C-c C-m "exit" C-m
    done
fi
echo "tmux has stopped."