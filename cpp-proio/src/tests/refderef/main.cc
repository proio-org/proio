#undef NDEBUG

#include <assert.h>
#include <stdio.h>
#include <iostream>

#include "event.h"
#include "lcio.pb.h"
#include "reader.h"
#include "writer.h"

using namespace proio::model;

void refderef1() {
    char filename[] = "refderef1XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));

    auto event = new proio::Event();
    auto parent = new lcio::MCParticle();
    parent->set_pdg(443);
    auto parentID = event->AddEntry(parent, "MCParticles");
    auto child1 = new lcio::MCParticle();
    child1->set_pdg(11);
    auto child1ID = event->AddEntry(child1, "MCParticles");
    auto child2 = new lcio::MCParticle();
    child2->set_pdg(-11);
    auto child2ID = event->AddEntry(child2, "MCParticles");
    parent->add_children(child1ID);
    parent->add_children(child2ID);
    child1->add_parents(parentID);
    child2->add_parents(parentID);

    auto parent_orig = parent->New();
    parent_orig->CopyFrom(*parent);
    auto child1_orig = child1->New();
    child1_orig->CopyFrom(*child1);
    auto child2_orig = child2->New();
    child2_orig->CopyFrom(*child2);

    writer->Push(event);

    delete writer;

    auto reader = new proio::Reader(filename);

    auto event_ = reader->Next();
    assert(event_);

    auto mcParticles = event_->TaggedEntries("MCParticles");
    auto parent_ = (lcio::MCParticle *)event_->GetEntry(mcParticles[0]);
    assert(parent_->DebugString().compare(parent_orig->DebugString()) == 0);
    auto child1_ = (lcio::MCParticle *)event_->GetEntry(parent_->children()[0]);
    assert(child1_->DebugString().compare(child1_orig->DebugString()) == 0);
    auto child2_ = (lcio::MCParticle *)event_->GetEntry(parent_->children()[1]);
    assert(child2_->DebugString().compare(child2_orig->DebugString()) == 0);
    parent_ = (lcio::MCParticle *)event_->GetEntry(child1_->parents()[0]);
    assert(parent_->DebugString().compare(parent_orig->DebugString()) == 0);
    parent_ = (lcio::MCParticle *)event_->GetEntry(child2_->parents()[0]);
    assert(parent_->DebugString().compare(parent_orig->DebugString()) == 0);

    delete reader;
    remove(filename);
    delete event;
    delete event_;
    delete parent_orig;
    delete child1_orig;
    delete child2_orig;
}

void refderef2() {
    auto event = new proio::Event();
    auto parent = new lcio::MCParticle();
    parent->set_pdg(443);
    auto parentID = event->AddEntry(parent, "MCParticles");
    auto child1 = new lcio::MCParticle();
    child1->set_pdg(11);
    auto child1ID = event->AddEntry(child1, "MCParticles");
    auto child2 = new lcio::MCParticle();
    child2->set_pdg(-11);
    auto child2ID = event->AddEntry(child2, "MCParticles");
    parent->add_children(child1ID);
    parent->add_children(child2ID);
    child1->add_parents(parentID);
    child2->add_parents(parentID);

    auto mcParticles = event->TaggedEntries("MCParticles");
    auto parent_ = (lcio::MCParticle *)event->GetEntry(mcParticles[0]);
    assert(parent_->DebugString().compare(parent->DebugString()) == 0);
    auto child1_ = (lcio::MCParticle *)event->GetEntry(parent_->children()[0]);
    assert(child1_->DebugString().compare(child1->DebugString()) == 0);
    auto child2_ = (lcio::MCParticle *)event->GetEntry(parent_->children()[1]);
    assert(child2_->DebugString().compare(child2->DebugString()) == 0);
    parent_ = (lcio::MCParticle *)event->GetEntry(child1_->parents()[0]);
    assert(parent_->DebugString().compare(parent->DebugString()) == 0);
    parent_ = (lcio::MCParticle *)event->GetEntry(child2_->parents()[0]);
    assert(parent_->DebugString().compare(parent->DebugString()) == 0);

    delete event;
}

int main() {
    refderef1();
    refderef2();
}
