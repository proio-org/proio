#include <iostream>

#include "event.h"

eicio::Event::Event() { std::cout << "Constructor!" << std::endl; }

eicio::Event::~Event() { std::cout << "Destructor!" << std::endl; }
