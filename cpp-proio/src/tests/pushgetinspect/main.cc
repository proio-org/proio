#include <stdio.h>
#include <iostream>

#include "event.h"
#include "lcio.pb.h"
#include "reader.h"
#include "writer.h"

void pushGetInspect1(proio::Compression comp) {
    char filename[] = "pushgetinspectXXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));
    writer->SetCompression(comp);

    auto event = new proio::Event();
    auto part = new proio::model::lcio::MCParticle();
    part->set_pdg(11);
    event->AddEntry("MCParticles", part);
    part = new proio::model::lcio::MCParticle();
    part->set_pdg(-11);
    event->AddEntry("MCParticles", part);
    writer->Push(event);
    writer->Push(event);
    delete event;

    delete writer;

    auto reader = new proio::Reader(filename);

    while ((event = reader->Next())) {
        auto ids = event->TaggedEntries("MCParticles");
        for (auto id : ids) {
            std::cout << event->GetEntry(id)->DebugString() << std::flush;
        }
        delete event;
    }

    delete reader;
    remove(filename);
}

int main() { pushGetInspect1(proio::UNCOMPRESSED); }
