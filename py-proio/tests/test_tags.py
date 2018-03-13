import pytest

import proio
import proio.model.eic as eic

def test_tag_cleanup1():
    event = proio.Event()
    entry = eic.SimHit()
    entryID = event.add_entry('SimCreated', entry)
    assert(len(event.tagged_entries('SimCreated')) == 1)
    assert(len(event.all_entries()) == 1)
    event.remove_entry(entryID)
    assert(len(event.tagged_entries('SimCreated')) == 0)
    assert(len(event.all_entries()) == 0)

def test_tag_cleanup2():
    event = proio.Event()
    entry = eic.SimHit()
    entryID = event.add_entry('SimCreated', entry)
    event._flush_cache()
    assert(len(event.tagged_entries('SimCreated')) == 1)
    assert(len(event.all_entries()) == 1)
    event.remove_entry(entryID)
    assert(len(event.tagged_entries('SimCreated')) == 0)
    assert(len(event.all_entries()) == 0)

def test_untag1():
    event = proio.Event()
    entry = eic.SimHit()
    entryID = event.add_entry('SimCreated', entry)
    assert(len(event.tagged_entries('SimCreated')) == 1)
    event.untag_entry(entryID, 'WrongTag')
    assert(len(event.tagged_entries('SimCreated')) == 1)
    assert(len(event.tags()) == 1)
    assert(len(event.all_entries()) == 1)
    event.untag_entry(entryID, 'SimCreated')
    assert(len(event.tagged_entries('SimCreated')) == 0)
    assert(len(event.all_entries()) == 1)

def test_untag2():
    event = proio.Event()
    entry = eic.SimHit()
    entryID = event.add_entry('SimCreated', entry)
    assert(len(event.tagged_entries('SimCreated')) == 1)
    event.untag_entry(entryID + 1, 'WrongTag')
    assert(len(event.tagged_entries('SimCreated')) == 1)
    assert(len(event.tags()) == 1)
    assert(len(event.all_entries()) == 1)
    event.untag_entry(entryID + 1, 'SimCreated')
    assert(len(event.tagged_entries('SimCreated')) == 1)
    assert(len(event.all_entries()) == 1)

def test_tag_delete1():
    event = proio.Event()
    entry = eic.SimHit()
    entryID = event.add_entry('SimCreated', entry)
    assert(len(event.tags()) == 1)
    assert(len(event.all_entries()) == 1)
    event.delete_tag('WrongTag')
    assert(len(event.tags()) == 1)
    event.delete_tag('SimCreated')
    assert(len(event.tags()) == 0)
    assert(len(event.all_entries()) == 1)

def test_tag_rev_lookup1():
    event = proio.Event()
    entry = eic.SimHit()
    entryID = event.add_entry('SimCreated', entry)
    assert(len(event.entry_tags(entryID)) == 1)
    assert(event.entry_tags(entryID)[0] == 'SimCreated')
    assert(len(event.tags()) == 1)
    event.tag_entry(entryID, 'Tracker')
    assert(len(event.entry_tags(entryID)) == 2)
    assert(len(event.tags()) == 2)
    event.untag_entry(entryID, 'SimCreated')
    assert(len(event.entry_tags(entryID)) == 1)
    assert(event.entry_tags(entryID)[0] == 'Tracker')
    assert(len(event.tags()) == 2)

