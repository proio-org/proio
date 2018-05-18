#undef NDEBUG

#include <assert.h>
#include <stdio.h>
#include <iostream>

#include "eic.pb.h"
#include "event.h"
#include "reader.h"
#include "writer.h"

namespace model = proio::model::eic;

void unsetValue1() {
    char filename[] = "unsetvalue1.XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));

    auto event = new proio::Event();
    auto id = event->AddEntry(new model::Particle());
    writer->Push(event);
    delete event;

    delete writer;
    auto reader = new proio::Reader(filename);

    event = reader->Next();
    auto part = event->GetEntry(id);
    auto desc = part->GetDescriptor();
    auto fieldDesc = desc->FindFieldByName("charge");
    auto refl = part->GetReflection();
    assert(!refl->HasField(*part, fieldDesc));
    assert(((model::Particle *)part)->charge() == 0);

    delete reader;
    remove(filename);
}

void setValue1() {
    char filename[] = "setvalue1.XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));

    auto event = new proio::Event();
    auto partOut = new model::Particle();
    partOut->set_charge(0);
    auto id = event->AddEntry(partOut);
    writer->Push(event);
    delete event;

    delete writer;
    auto reader = new proio::Reader(filename);

    event = reader->Next();
    auto part = event->GetEntry(id);
    auto desc = part->GetDescriptor();
    auto fieldDesc = desc->FindFieldByName("charge");
    auto refl = part->GetReflection();
    assert(refl->HasField(*part, fieldDesc));
    assert(((model::Particle *)part)->charge() == 0);

    delete reader;
    remove(filename);
}

int main() {
    unsetValue1();
    setValue1();
}
