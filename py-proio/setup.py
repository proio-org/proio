import glob
import os
import setuptools
import shutil
import subprocess
import sys

from distutils.command.clean import clean as _clean
from distutils.spawn import find_executable
from os.path import basename
from os.path import splitext

if sys.version_info[0] == 3:
    from distutils.command.build_py import build_py_2to3 as _build_py
else:
    from distutils.command.build_py import build_py as _build_py

models = [splitext(basename(model))[0] for model in glob.glob('../model/*.proto')]

protoc = find_executable('protoc')

def generate_main_proto():
    main_proto_path = 'proio/proto/proio_pb2.py'
    if os.path.exists(main_proto_path):
        return

    if protoc is None:
        print('protoc not found')
        sys.exit(1)

    protoc_command = [
    protoc,
        '--proto_path=proio/proto=../proto',
        '--python_out=.',
        '../proto/proio.proto',
        ]
    if subprocess.call(protoc_command) != 0:
        print('failed to call protoc for proio.proto')
        sys.exit(1)

def generate_model_proto(model):
    if os.path.exists('proio/model/' + model):
        return

    protoc_command = [
            protoc,
            '--proto_path=proio/model=../model'
            ,'--python_out=.',
            '../model/%s.proto' % model,
            ]
    if subprocess.call(protoc_command) != 0:
        print('failed to call protoc for model: ' + model)
        sys.exit(1)

    os.mkdir('proio/model/' + model)
    gen_file = model + '_pb2.py'
    os.rename('proio/model/' + gen_file, 'proio/model/%s/%s' % (model, gen_file))
    with open('proio/model/' + model + '/__init__.py', 'w') as init_file:
        init_file.write('from .%s_pb2 import *' % model)

class clean(_clean):
    def run(self):
        try:
            os.remove('proio/proto/proio_pb2.py')
        except OSError:
            pass

        for model in models:
            shutil.rmtree('proio/model/' + model, ignore_errors=True)

        _clean.run(self);

class build_py(_build_py):
    def run(self):
        generate_main_proto()

        for model in models:
            generate_model_proto(model)

        _build_py.run(self)

setuptools.setup(
        name = 'proio',
        version = '0.9',
        description = 'Library for reading and writing proio files and streams',
        url = 'http://github.com/decibelcooper/proio',
        author = 'David Blyth',
        author_email = 'dblyth@anl.gov',
        license = 'BSD-3-Clause',
        packages = setuptools.find_packages(),
        install_requires = ['protobuf', 'lz4==2.*'],
        zip_safe = True,
        cmdclass = {
            'clean': clean,
            'build_py' : build_py,
            }
      )
