#include <unistd.h>
#include <iostream>

#include "proio/model/lcio/lcio.pb.h"
#include "proio/reader.h"
#include "proio/writer.h"

using namespace proio;
using namespace proio::model;

int main(int argc, char **argv) {
    auto event = new Event;
    auto mcpColl = new lcio::MCParticleCollection;
    event->Add(mcpColl, "MCParticles");

    lcio::MCParticle *mcp0 = mcpColl->add_entries();
    mcp0->set_charge(1);
    mcp0->set_pdg(123456);
    lcio::MCParticle *mcp1 = mcpColl->add_entries();
    mcp1->set_charge(1);
    event->Reference(mcp0, mcp1->add_parents());
    std::string dbs0(event->Dereference(mcp1->parents(0))->DebugString());

    auto writer = new Writer(".test.proio.tmp");
    writer->Push(event);
    delete event;
    delete writer;

    auto reader = new Reader(".test.proio.tmp");
    event = reader->Get();
    mcpColl = (lcio::MCParticleCollection *)event->Get("MCParticles");
    mcp1 = mcpColl->mutable_entries(1);
    std::string dbs1(event->Dereference(mcp1->parents(0))->DebugString());
    delete reader;

    if (dbs1.compare(dbs0) != 0) return EXIT_FAILURE;

    return EXIT_SUCCESS;
}
