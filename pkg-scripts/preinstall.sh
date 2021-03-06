#!/bin/sh

# check that owner group exists
if [ -z `getent group frontman` ]; then
  groupadd frontman
fi

# check that user exists
if [ -z `getent passwd frontman` ]; then
  useradd  --gid frontman --system --shell /bin/false frontman
fi

# remove deprecated sysctl setting
if [ -e /etc/sysctl.d/50-ping_group_range.conf ]; then
  rm -f /etc/sysctl.d/50-ping_group_range.conf
fi
