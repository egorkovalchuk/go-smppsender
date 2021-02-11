#!/bin/bash

PID_FILE="smsc.pid"
APP_PID=

log()
{
    shift
    echo "$*"
}

function stop()
{
    if [ -f $PID_FILE ]; then
    read PID<$PID_FILE
    kill ${PID}
    fi
    echo "Daemon stop"
}

function exit_rm()
{
    if [ -f $PID_FILE ]; then
        rm $PID_FILE
        echo "Daemon stop"
    fi
    exit 1
}

start()
{

    nohup ./smsc -d -t &
    APP_PID=$!

    if [ -f $PID_FILE ]; then
    read PID<$PID_FILE

        if [[ -n ps aux | grep $PID | grep -v grep  ]]; then
            echo "Daemon working. Stop start"
            exit
        else
            echo $APP_PID > $PID_FILE
        fi
    else
      echo $APP_PID > $PID_FILE
    fi

#Обработка прерывания
trap exit_rm SIGINT
trap exit_rm SIGKILL
}

case $1 in
    "start")
        start
    ;;
    "stop")
        stop
    ;;
    "help")
      ./smsc -h
    ;;
    )
        echo "$0 (start|stop|help)"
    ;;
esac
exit