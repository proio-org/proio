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

void pushGetInspect2(proio::Compression comp) {
    char filename[] = "pushgetinspect2XXXXXX";
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
    writer->Flush();

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

void pushSkipGet1(proio::Compression comp) {
    char filename[] = "pushskipget1XXXXXX";
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

    reader->Skip(1);
    event = reader->Next();
    assert(event->String().compare(eventsOut[1]->String()) == 0);
    delete event;
    for (int i = 0; i < eventsOut.size(); i++) delete eventsOut[i];

    delete reader;
    remove(filename);
}

void pushSkipGet2(proio::Compression comp) {
    char filename[] = "pushskipget2XXXXXX";
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
    writer->Flush();

    event = new proio::Event();
    event->AddEntry("TrackerHits", new lcio::SimTrackerHit());
    event->AddEntry("TrackerHits", new lcio::SimTrackerHit());
    writer->Push(event);
    eventsOut.push_back(event);

    delete writer;

    auto reader = new proio::Reader(filename);

    reader->Skip(1);
    event = reader->Next();
    assert(event->String().compare(eventsOut[1]->String()) == 0);
    delete event;
    for (int i = 0; i < eventsOut.size(); i++) delete eventsOut[i];

    delete reader;
    remove(filename);
}

int main() {
    pushGetInspect1(proio::LZ4);
    pushGetInspect1(proio::UNCOMPRESSED);
    pushGetInspect1(proio::GZIP);

    pushGetInspect2(proio::LZ4);
    pushGetInspect2(proio::UNCOMPRESSED);
    pushGetInspect2(proio::GZIP);

    pushSkipGet1(proio::LZ4);
    pushSkipGet1(proio::UNCOMPRESSED);
    pushSkipGet1(proio::GZIP);

    pushSkipGet2(proio::LZ4);
    pushSkipGet2(proio::UNCOMPRESSED);
    pushSkipGet2(proio::GZIP);
}
