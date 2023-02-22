session_name=dousheng

tmux has-session -t $session_name
if [ $? != 0 ];then
    path_to_scrpit=scripts/microservice
    cd $path_to_scrpit

    tmux new-session -s $session_name -n comment -d
    tmux send-keys -t $session_name 'sh comment.sh' C-m
    tmux new-window -n favorite -t $session_name
    tmux send-keys -t $session_name:1 'sh favorite.sh' C-m
    tmux new-window -n message -t $session_name
    tmux send-keys -t $session_name:2 'sh message.sh' C-m
    tmux new-window -n relation -t $session_name
    tmux send-keys -t $session_name:3 'sh relation.sh' C-m
    tmux new-window -n user -t $session_name
    tmux send-keys -t $session_name:4 'sh user.sh' C-m
    tmux new-window -n video -t $session_name
    tmux send-keys -t $session_name:5 'sh video.sh' C-m
    tmux new-window -n api -t $session_name
    tmux send-keys -t $session_name:6 'sh api.sh' C-m
    tmux select-window -t $session_name:6
fi
tmux attach -t dousheng
echo "tmux has started."