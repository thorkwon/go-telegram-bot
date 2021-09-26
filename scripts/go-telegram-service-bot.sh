#!/usr/bin/env bash

function help()
{
	echo "Usage: $0 command"
	echo "Command:"
	echo "  status		get service status"
	echo "  restart		service restart"
	echo "  stop			service stop"
	echo "  start			service start"
	echo "  log [line]		get service log"
	exit 1
}

if [ $# == 0 ]; then
	help
fi

case $1 in
	status)
		systemctl status go-telegram-service-bot
		;;
	restart)
		sudo systemctl restart go-telegram-service-bot
		;;
	stop)
		sudo systemctl stop go-telegram-service-bot
		;;
	start)
		sudo systemctl start go-telegram-service-bot
		;;
	log)
		echo -e "Press output log: ^c\n"
		if [ -z $2 ]; then
			tail -f ~/.config/go-telegram-bot/bot.log
		else
			tail -n $2 -f ~/.config/go-telegram-bot/bot.log
		fi
		;;
	*)
		help
		;;
esac
