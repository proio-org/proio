#include <unistd.h>
#include <iostream>

#include "eicio/reader.h"
#include "eicio/writer.h"

using namespace eicio;

int main(int argc, char **argv) {
    auto event = new Event;
    auto mcpColl = new model::MCParticleCollection;
    event->Add(mcpColl, "MCParticles");

    model::MCParticle *mcp0 = mcpColl->add_entries();
    mcp0->set_charge(1);
    mcp0->set_pdg(123456);
    model::MCParticle *mcp1 = mcpColl->add_entries();
    mcp1->set_charge(1);
    event->Reference(mcp0, mcp1->add_parents());
    std::string dbs0(event->Dereference(mcp1->parents(0))->DebugString());

    auto writer = new Writer(".test.eicio.tmp");
    writer->Push(event);
    delete event;
    delete writer;

    auto reader = new Reader(".test.eicio.tmp");
    event = reader->Get();
    mcpColl = (model::MCParticleCollection *)event->Get("MCParticles");
    mcp1 = mcpColl->mutable_entries(1);
    std::string dbs1(event->Dereference(mcp1->parents(0))->DebugString());
    delete reader;

    if (dbs1.compare(dbs0) != 0) return EXIT_FAILURE;

    return EXIT_SUCCESS;
}
