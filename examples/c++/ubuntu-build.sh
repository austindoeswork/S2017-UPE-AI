#!/bin/bash
g++ -c easywsclient.cpp -o easywsclient.o
g++ -c ai-ubuntu.cpp -o ai.o
g++ ai.o easywsclient.o -o ai
rm ai.o
rm easywsclient.o
