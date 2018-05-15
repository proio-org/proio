#undef NDEBUG

#include <assert.h>
#include <stdio.h>
#include <iostream>

#include "eic.pb.h"
#include "event.h"
#include "reader.h"
#include "writer.h"

using namespace proio::model;

void tagInspect1() {
    auto event = new proio::Event();

    auto id = event->AddEntry(new eic::Particle(), "Tag1");
    event->TagEntry(id, "Tag2");

    auto tags = event->EntryTags(id);
    assert(tags[0].compare("Tag1") == 0);
    assert(tags[1].compare("Tag2") == 0);

    delete event;
}

void tagInspect2() {
    char filename[] = "tagInspect2XXXXXX";
    auto writer = new proio::Writer(mkstemp(filename));

    auto event = new proio::Event();
    auto id = event->AddEntry(new eic::Particle(), "Tag1");
    event->TagEntry(id, "Tag2");
    writer->Push(event);
    delete event;

    delete writer;

    auto reader = new proio::Reader(filename);

    event = reader->Next();
    auto tags = event->EntryTags(id);
    assert(tags[0].compare("Tag1") == 0);
    assert(tags[1].compare("Tag2") == 0);

    delete reader;
    remove(filename);
}

void tagRemoveInspect1() {
    auto event = new proio::Event();

    auto id = event->AddEntry(new eic::Particle(), "Tag1");
    event->TagEntry(id, "Tag2");
    event->DeleteTag("Tag1");

    auto tags = event->EntryTags(id);
    assert(tags.size() == 1);
    assert(tags[0].compare("Tag2") == 0);

    delete event;
}

void tagRemoveInspect2() {
    auto event = new proio::Event();

    auto id = event->AddEntry(new eic::Particle(), "Tag1");
    event->TagEntry(id, "Tag2");
    event->UntagEntry(id, "Tag1");

    auto tags = event->EntryTags(id);
    assert(tags.size() == 1);
    assert(tags[0].compare("Tag2") == 0);

    delete event;
}

void entryRemoveTagInspect1() {
    auto event = new proio::Event();

    auto id1 = event->AddEntry(new eic::Particle(), "Tag1");
    event->TagEntry(id1, "Tag2");
    auto id2 = event->AddEntry(new eic::Particle(), "Tag1");
    event->TagEntry(id2, "Tag2");
    event->RemoveEntry(id1);

    auto tag1Entries = event->TaggedEntries("Tag1");
    assert(tag1Entries.size() == 1);
    assert(tag1Entries[0] == id2);
    auto tag2Entries = event->TaggedEntries("Tag2");
    assert(tag2Entries.size() == 1);
    assert(tag2Entries[0] == id2);

    assert(event->AllEntries().size() == 1);

    delete event;
}

int main() {
    tagInspect1();
    tagInspect2();
    tagRemoveInspect1();
    tagRemoveInspect2();
    entryRemoveTagInspect1();
}
