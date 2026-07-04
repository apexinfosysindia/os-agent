#!/bin/bash
systemctl daemon-reload
systemctl enable apexos-agent
systemctl start apexos-agent
