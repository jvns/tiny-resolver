#!/bin/bash

get_record() {
    grep -E "\s$1\s" | head -n 1 | awk '{print $5}'
}

lookup () {
    echo dig -r "$@" >&2
    dig -r +norecurse +noall +authority +answer +additional "$@"
}

resolve() {
   DOMAIN="$1"
   # start with a `.` nameserver. That's easy.
   NAMESERVER="198.41.0.4"
   while true
   do
       RESPONSE=$(lookup @"$NAMESERVER" "$DOMAIN")
       IP=$(echo "$RESPONSE" | grep "$DOMAIN" | get_record "A" )
       GLUEIP=$(echo "$RESPONSE" | get_record "A" | grep -v "$DOMAIN")
       NS=$(echo "$RESPONSE" | get_record "NS")
       if [ -n "$IP" ]; then
           echo "$IP"
           return 0
       elif [ -n "$GLUEIP" ]; then
           NAMESERVER="$GLUEIP"
       elif [ -n "$NS" ]; then
           NAMESERVER=$(resolve "$NS")
       else
           echo "No IP found for $DOMAIN"
           exit 1
       fi
   done
}

resolve "$1"
