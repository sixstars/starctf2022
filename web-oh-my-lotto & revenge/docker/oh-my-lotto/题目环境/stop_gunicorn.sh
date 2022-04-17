#!/usr/bin/env bash

kill $(lsof -i:6680|awk '{if(NR==2)print $2}')