#!/bin/sh
/app/auth_service &
nginx -g 'daemon off;'