import io
import pytest

import proio
import proio.model.lcio as prolcio

def test_ref_deref1():
    buf = io.BytesIO(b'')
    with proio.Writer(fileobj = buf) as writer:
        event = proio.Event()
        parent = prolcio.MCParticle()
        parent.PDG = 443
        parent_id = event.add_entry('MCParticles', parent)
        child1 = prolcio.MCParticle()
        child1.PDG = 11
        child2 = prolcio.MCParticle()
        child2.PDG = -11
        child_ids = event.add_entries('MCParticles', child1, child2)
        parent.children.extend(child_ids)
        child1.parents.append(parent_id)
        child2.parents.append(parent_id)

        writer.push(event)

    buf.seek(0, 0)
    
    with proio.Reader(fileobj = buf) as reader:
        event = reader.__next__()
        assert event != None

        mc_particles = event.tagged_entries('MCParticles')
        assert mc_particles != None

        parent_ = event.get_entry(mc_particles[0])
        assert parent_.__str__() == parent.__str__()
        child1_ = event.get_entry(parent_.children[0])
        assert child1_.__str__() == child1.__str__()
        child2_ = event.get_entry(parent_.children[1])
        assert child2_.__str__() == child2.__str__()
        parent_ = event.get_entry(child1_.parents[0])
        assert parent_.__str__() == parent.__str__()
        parent_ = event.get_entry(child2_.parents[0])
        assert parent_.__str__() == parent.__str__()

def test_ref_deref2():
    event = proio.Event()
    parent = prolcio.MCParticle()
    parent.PDG = 443
    parent_id = event.add_entry('MCParticles', parent)
    child1 = prolcio.MCParticle()
    child1.PDG = 11
    child2 = prolcio.MCParticle()
    child2.PDG = -11
    child_ids = event.add_entries('MCParticles', child1, child2)
    parent.children.extend(child_ids)
    child1.parents.append(parent_id)
    child2.parents.append(parent_id)

    mc_particles = event.tagged_entries('MCParticles')
    assert mc_particles != None

    parent_ = event.get_entry(mc_particles[0])
    assert parent_.__str__() == parent.__str__()
    child1_ = event.get_entry(parent_.children[0])
    assert child1_.__str__() == child1.__str__()
    child2_ = event.get_entry(parent_.children[1])
    assert child2_.__str__() == child2.__str__()
    parent_ = event.get_entry(child1_.parents[0])
    assert parent_.__str__() == parent.__str__()
    parent_ = event.get_entry(child2_.parents[0])
    assert parent_.__str__() == parent.__str__()

def test_get_bad_entry1():
    event = proio.Event()
    assert(event.get_entry(1) is None)
