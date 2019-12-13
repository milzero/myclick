# -*- coding:utf-8 -*-
import requests

def ip2uint(ip):
    numbs = ip.split('.')
    if len(numbs) != 4:
        return -1
    numbs = [int(i) for i in numbs if int(i) >= 0 or int(i) < 256]
    if len(numbs) != 4:
        return -1
    numb = 0xFF & numbs[0]
    numb = (numb << 8)+numbs[1]
    numb = (numb << 8)+numbs[2]
    numb = (numb << 8)+numbs[3]
    return numb


def uint2ip(numb):
    ip = list()
    ip.append(str(numb & 0xff))
    ip.append(str((numb >> 8) & 0xff))
    ip.append(str((numb >> 16) & 0xff))
    ip.append(str((numb >> 24) & 0xff))
    ip.reverse()
    return ".".join(ip)

def iplocation(ip):
    pass
