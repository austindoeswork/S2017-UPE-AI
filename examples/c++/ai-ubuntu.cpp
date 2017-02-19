// USES: https://github.com/dhbaird/easywsclient
// built on Ubuntu 16.10

/*
  COMPILATION/RUNNING INSTRUCTIONS:
  g++ -c easywsclient.cpp -o easywsclient.o
  g++ -c ai-ubuntu.cpp -o ai.o
  g++ ai.o easywsclient.o -o ai
  ./ai
*/

#include "easywsclient.hpp"
#include <iostream>
#include <string>

#include <assert.h> // totally optional

void handle_message(const std::string & message)
{
  std::cout << message << std::endl;
}

int main() {
  std::string serverURL = "ws://npcompete.io";
  std::string devkey = "TrumansLoudlySquareBellybutton";
  std::string input;

  // figure out room name, edit serverURL if necessary
  std::cout << "Enter game lobby name (or blank to enter matchmaking): ";
  std::getline(std::cin, input);

  if (input.size() > 0) {
    serverURL += "/wsjoin?game=" + input;
  } else {
    serverURL += "/wsplay";
  }

  // init client
  using easywsclient::WebSocket;
  WebSocket::pointer ws = WebSocket::from_url(serverURL);
  assert(ws);
  
  // first message we send is the devkey
  ws->send(devkey);
  while (ws->getReadyState() != WebSocket::CLOSED) {
    ws->poll();
    ws->send("b00 01");
    ws->dispatch(handle_message);
  }
  delete ws; // alternatively, use unique_ptr<> if you have C++11
  return 0;
}
