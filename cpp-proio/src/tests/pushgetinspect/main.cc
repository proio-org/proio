#undef NDEBUG

#include <assert.h>
#include <stdio.h>
#include <iostream>

#include "eic.pb.h"
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
    event->AddEntry(new lcio::MCParticle(), "MCParticles");
    event->AddEntry(new lcio::MCParticle(), "MCParticles");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    writer->Push(event);
    eventsOut.push_back(event);

    event = new proio::Event();
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    writer->Push(event);
    eventsOut.push_back(event);

    delete writer;

    auto reader = new proio::Reader(filename);

    for (int i = 0; i < eventsOut.size(); i++) {
        event = reader->Next();
        assert(event);
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
    event->AddEntry(new lcio::MCParticle(), "MCParticles");
    event->AddEntry(new lcio::MCParticle(), "MCParticles");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    writer->Push(event);
    eventsOut.push_back(event);
    writer->Flush();

    event = new proio::Event();
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    writer->Push(event);
    eventsOut.push_back(event);

    delete writer;

    auto reader = new proio::Reader(filename);

    for (int i = 0; i < eventsOut.size(); i++) {
        event = reader->Next();
        assert(event);
        assert(event->String().compare(eventsOut[i]->String()) == 0);
        delete eventsOut[i];
        delete event;
    }

    delete reader;
    remove(filename);
}

void pushGetInspect3(proio::Compression comp) {
    char filename[] = "pushgetinspect3XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));
    writer->SetCompression(comp);

    std::vector<proio::Event *> eventsOut;

    auto event = new proio::Event();
    event->AddEntry(new lcio::MCParticle(), "MCParticles");
    event->AddEntry(new lcio::MCParticle(), "MCParticles");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    auto eventOut = new proio::Event();
    *eventOut = *event;
    eventsOut.push_back(eventOut);
    writer->Push(event);
    writer->Flush();

    event->Clear();
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    writer->Push(event);
    eventOut = new proio::Event();
    *eventOut = *event;
    eventsOut.push_back(eventOut);

    delete writer;

    auto reader = new proio::Reader(filename);

    for (int i = 0; i < eventsOut.size(); i++) {
        assert(reader->Next(event));
        assert(event->String().compare(eventsOut[i]->String()) == 0);
        delete eventsOut[i];
    }

    delete reader;
    delete event;
    remove(filename);
}

void pushGetInspect4(proio::Compression comp) {
    char filename[] = "pushgetinspect4XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));
    writer->SetCompression(comp);

    std::vector<proio::Event *> eventsOut;

    auto event = new proio::Event();
    auto partDesc =
        google::protobuf::DescriptorPool::generated_pool()->FindMessageTypeByName("proio.model.eic.Particle");
    auto part = (eic::Particle *)event->Free(partDesc);
    assert(!part);
    event->AddEntry(new eic::Particle(), "Particle");
    event->AddEntry(new eic::Particle(), "Particle");
    auto hitDesc =
        google::protobuf::DescriptorPool::generated_pool()->FindMessageTypeByName("proio.model.eic.SimHit");
    auto hit = (eic::SimHit *)event->Free(hitDesc);
    assert(!hit);
    event->AddEntry(new eic::SimHit(), "Tracker");
    event->AddEntry(new eic::SimHit(), "Tracker");
    auto eventOut = new proio::Event();
    *eventOut = *event;
    eventsOut.push_back(eventOut);
    writer->Push(event);
    writer->Flush();

    event->Clear();
    hit = (eic::SimHit *)event->Free(hitDesc);
    assert(hit);
    event->AddEntry(hit, "Tracker");
    hit = (eic::SimHit *)event->Free(hitDesc);
    assert(hit);
    event->AddEntry(hit, "Tracker");
    hit = (eic::SimHit *)event->Free(hitDesc);
    assert(!hit);
    writer->Push(event);
    eventOut = new proio::Event();
    *eventOut = *event;
    eventsOut.push_back(eventOut);
    event->Clear();
    hit = (eic::SimHit *)event->Free(hitDesc);
    assert(hit);
    delete hit;

    delete writer;

    auto reader = new proio::Reader(filename);

    for (int i = 0; i < eventsOut.size(); i++) {
        assert(reader->Next(event));
        assert(event->String().compare(eventsOut[i]->String()) == 0);
        delete eventsOut[i];
    }

    delete reader;
    delete event;
    remove(filename);
}

void pushSkipGet1(proio::Compression comp) {
    char filename[] = "pushskipget1XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));
    writer->SetCompression(comp);

    std::vector<proio::Event *> eventsOut;

    auto event = new proio::Event();
    event->AddEntry(new lcio::MCParticle(), "MCParticles");
    event->AddEntry(new lcio::MCParticle(), "MCParticles");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    writer->Push(event);
    eventsOut.push_back(event);

    event = new proio::Event();
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    writer->Push(event);
    eventsOut.push_back(event);

    delete writer;

    auto reader = new proio::Reader(filename);

    reader->Skip(1);
    event = reader->Next();
    assert(event);
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
    event->AddEntry(new lcio::MCParticle(), "MCParticles");
    event->AddEntry(new lcio::MCParticle(), "MCParticles");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    writer->Push(event);
    eventsOut.push_back(event);
    writer->Flush();

    event = new proio::Event();
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    writer->Push(event);
    eventsOut.push_back(event);

    delete writer;

    auto reader = new proio::Reader(filename);

    reader->Skip(1);
    event = reader->Next();
    assert(event);
    assert(event->String().compare(eventsOut[1]->String()) == 0);
    delete event;
    for (int i = 0; i < eventsOut.size(); i++) delete eventsOut[i];

    delete reader;
    remove(filename);
}

void pushSkipGet3(proio::Compression comp) {
    char filename[] = "pushskipget3XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));
    writer->SetCompression(comp);

    std::vector<proio::Event *> eventsOut;

    auto event = new proio::Event();
    event->AddEntry(new lcio::MCParticle(), "MCParticles");
    event->AddEntry(new lcio::MCParticle(), "MCParticles");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    writer->Push(event);
    writer->Push(event);
    auto eventOut = new proio::Event();
    *eventOut = *event;
    eventsOut.push_back(eventOut);
    writer->Flush();

    event->Clear();
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    event->AddEntry(new lcio::SimTrackerHit(), "TrackerHits");
    writer->Push(event);
    eventOut = new proio::Event();
    *eventOut = *event;
    eventsOut.push_back(eventOut);

    delete writer;

    auto reader = new proio::Reader(filename);

    assert(reader->Skip(10) == 3);
    reader->SeekToStart();
    assert(reader->Skip(1) == 1);
    for (int i = 0; i < eventsOut.size(); i++) {
        assert(reader->Next(event));
        assert(event->String().compare(eventsOut[i]->String()) == 0);
        delete eventsOut[i];
    }
    assert(reader->Skip(1) == 0);

    delete reader;
    delete event;
    remove(filename);
}

int main() {
    pushGetInspect1(proio::LZ4);
    pushGetInspect1(proio::UNCOMPRESSED);
    pushGetInspect1(proio::GZIP);

    pushGetInspect2(proio::LZ4);
    pushGetInspect2(proio::UNCOMPRESSED);
    pushGetInspect2(proio::GZIP);

    pushGetInspect3(proio::LZ4);
    pushGetInspect3(proio::UNCOMPRESSED);
    pushGetInspect3(proio::GZIP);

    pushGetInspect4(proio::LZ4);
    pushGetInspect4(proio::UNCOMPRESSED);
    pushGetInspect4(proio::GZIP);

    pushSkipGet1(proio::LZ4);
    pushSkipGet1(proio::UNCOMPRESSED);
    pushSkipGet1(proio::GZIP);

    pushSkipGet2(proio::LZ4);
    pushSkipGet2(proio::UNCOMPRESSED);
    pushSkipGet2(proio::GZIP);

    pushSkipGet3(proio::LZ4);
    pushSkipGet3(proio::UNCOMPRESSED);
    pushSkipGet3(proio::GZIP);
}
