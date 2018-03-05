from glob import glob
from os.path import basename
from os.path import splitext
from setuptools import setup

models = ['proio.model.' + splitext(basename(model))[0] for model in glob('proio/model/*')]
models = list(filter(lambda model: model != 'proio.model.__init__', models))

setup(name='proio',
      version='0.6.1',
      description='Library for reading and writing proio files and streams',
      url='http://github.com/decibelcooper/proio',
      author='David Blyth',
      author_email='dblyth@anl.gov',
      license='None',
      packages=['proio', 'proio.proto', 'proio.model'] + models,
      install_requires=['protobuf', 'lz4'],
      zip_safe=True)
