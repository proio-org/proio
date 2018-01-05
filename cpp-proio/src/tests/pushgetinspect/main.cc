#include <stdio.h>

#include "event.h"
#include "lcio.pb.h"
#include "writer.h"

int main() {
    char filename[] = "pushgetinspectXXXXXX";

    auto writer = new proio::Writer(mkstemp(filename));

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

    // remove(filename);
    return 0;
}
