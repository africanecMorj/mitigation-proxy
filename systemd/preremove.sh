#!/bin/sh

systemctl stop heavyrain.service || true
systemctl disable heavyrain.service || true
