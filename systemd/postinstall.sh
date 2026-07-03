#!/bin/sh
set -e

systemctl daemon-reload
systemctl enable heavyrain.service
