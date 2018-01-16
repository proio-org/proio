#undef NDEBUG

#include <assert.h>
#include <stdio.h>
#include <iostream>

#include "eic.pb.h"
#include "event.h"
#include "reader.h"
#include "writer.h"

using namespace proio::model;

void scan1(proio::Compression comp) {
    char filename[] = "scan1XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));
    writer->SetCompression(comp);

    std::vector<proio::Event *> eventsOut;

    auto event = new proio::Event();
    event->AddEntry(new eic::Particle());
    writer->Push(event);
    writer->Push(event);
    writer->Flush();
    writer->Push(event);
    eventsOut.push_back(event);
    writer->Push(event);
    eventsOut.push_back(event);
    writer->Push(event);
    eventsOut.push_back(event);

    delete writer;

    auto reader = new proio::Reader(filename);
    auto header = reader->NextHeader();
    assert(header->nevents() == 2);

    for (int i = 0; i < eventsOut.size(); i++) {
        auto event = reader->Next();
        assert(event);
        assert(event->String().compare(eventsOut[i]->String()) == 0);
        delete event;
    }

    delete reader;
    delete event;
    remove(filename);
}

int main() {
    scan1(proio::LZ4);
    scan1(proio::GZIP);
    scan1(proio::UNCOMPRESSED);
}
