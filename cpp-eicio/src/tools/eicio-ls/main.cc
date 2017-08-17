#include <iostream>

#include "eicio/reader.h"

int main(int argc, char **argv) {
    auto reader = new eicio::Reader(argv[1]);

    while (auto event = reader->Get()) {
        event->GetHeader()->PrintDebugString();

        for (auto name : event->GetNames()) {
            auto coll = event->Get(name);
            if (coll != NULL) {
                coll->PrintDebugString();
            }
        }

        delete event;
    }

    delete reader;
    return EXIT_SUCCESS;
}
