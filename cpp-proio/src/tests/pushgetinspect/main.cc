#include <stdio.h>

#include "event.h"
#include "writer.h"

int main() {
    char filename[] = "pushgetinspectXXXXXX";

    auto writer = new proio::Writer(mkstemp(filename));

    auto event = new proio::Event();
    writer->Push(event);
    delete event;

    delete writer;

    remove(filename);
    return 0;
}
