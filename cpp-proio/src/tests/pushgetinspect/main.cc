#undef NDEBUG

#include <assert.h>
#include <stdio.h>
#include <iostream>

#include "event.h"
#include "lcio.pb.h"
#include "reader.h"
#include "writer.h"

using namespace proio::model;

void pushGetInspect1(proio::Compression comp) {
    char filename[] = "pushgetinspect1XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));
    writer->SetCompression(comp);

    std::vector<proio::Event *> eventsOut;

    auto event = new proio::Event();
    event->AddEntry("MCParticles", new lcio::MCParticle());
    event->AddEntry("MCParticles", new lcio::MCParticle());
    event->AddEntry("TrackerHits", new lcio::SimTrackerHit());
    event->AddEntry("TrackerHits", new lcio::SimTrackerHit());
    writer->Push(event);
    eventsOut.push_back(event);

    event = new proio::Event();
    event->AddEntry("TrackerHits", new lcio::SimTrackerHit());
    event->AddEntry("TrackerHits", new lcio::SimTrackerHit());
    writer->Push(event);
    eventsOut.push_back(event);

    delete writer;

    auto reader = new proio::Reader(filename);

    for (int i = 0; i < eventsOut.size(); i++) {
        event = reader->Next();
        assert(event->String().compare(eventsOut[i]->String()) == 0);
        delete eventsOut[i];
        delete event;
    }

    delete reader;
    remove(filename);
}

int main() {
    pushGetInspect1(proio::LZ4);
    pushGetInspect1(proio::UNCOMPRESSED);
    pushGetInspect1(proio::GZIP);
}
