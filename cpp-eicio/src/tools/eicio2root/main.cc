#include <iostream>

#include "eicio/reader.h"

int main(int argc, char **argv) {
    auto reader = new eicio::Reader(argv[1]);

    int nEvents = 0;
    while (auto event = reader->Get()) {
        nEvents++;
        event->GetHeader()->PrintDebugString();
        delete event;
    }
    std::cout << nEvents << std::endl;

    delete reader;
    return EXIT_SUCCESS;
}
