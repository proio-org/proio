# proio for Python
## API
The API documentation is generated using Sphinx, and can be found
[here](https://decibelcooper.github.io/py-proio-docs/).

## Installing
Py-proio is maintained on [pipy](https://pypi.python.org/pypi/proio).  The
proio package can be installed via
`pip`:
```shell
pip install --user --upgrade proio
```

For information on what versions of Python are supported, please see the
[Travis CI page](https://travis-ci.org/decibelcooper/proio).

## Examples
### Push, get, inspect
```python
import proio
import proio.model.lcio as model

test_filename = 'test_file.proio'
with proio.Writer(test_filename) as writer:
    event = proio.Event()

    parent = model.MCParticle()
    parent.PDG = 443
    parent_id = event.add_entry('Particles', parent)
    event.tag_entry(parent_id, 'MC', 'Primary')

    child1 = model.MCParticle()
    child1.PDG = 11
    child2 = model.MCParticle()
    child2.PDG = -11
    child_ids = event.add_entries('Particles', child1, child2)
    for ID in child_ids:
        event.tag_entry(ID, 'MC', 'Simulated')

    parent.children.extend(child_ids)
    child1.parents.append(parent_id)
    child2.parents.append(parent_id)

    writer.push(event)
    
with proio.Reader(test_filename) as reader:
    event = reader.next()
    
    mc_parts = event.tagged_entries('Primary')
    print('%i Primary particle(s)...' % len(mc_parts))
    for i in range(0, len(mc_parts)):
        part = event.get_entry(mc_parts[i])
        print('%i. PDG: %i' % (i, part.PDG))

        print('  %i children...' % len(part.children))
        for j in range(0, len(part.children)):
            print('  %i. PDG: %i' % (j, event.get_entry(part.children[j]).PDG))
```

### Iterate
```python
import proio

n_events = 50
test_filename = 'test_file.proio'

with proio.Writer(test_filename) as writer:
    event = proio.Event()
    for i in range(0, n_events):
        writer.push(event)

n_events = 0

with proio.Reader(test_filename) as reader:
    for event in reader:
        n_events += 1

print(n_events)
```
