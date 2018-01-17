#!/usr/bin/env python
# -*- coding: utf-8 -*-


# CRON COMMAND:
# sudo crontab -e
# */5 * * * * python /home/dnssync/cron.py


import time, glob, os, datetime, shutil


DIRECTORY_PATH = "/home/alexis/dns-files/"


def deleteFiles():
    os.chdir(DIRECTORY_PATH)

    for file in glob.glob("*/*-*-*.*-*"):
        datestr = file[file.find("/")+1:]
        date = datetime.datetime.strptime(datestr,"%Y-%m-%d.%H-%M")

        if time.time() - time.mktime(date.timetuple()) > 60 * 60:
            shutil.rmtree(file)
            print "Removing directory: " + file


if __name__ == "__main__":
    deleteFiles()