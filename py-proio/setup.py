from glob import glob
from os.path import basename
from os.path import splitext
from setuptools import setup

modules = ['proio.model.' + splitext(basename(i))[0] for i in glob("proio/model/*.py")]

setup(name='proio',
      version='0.3.0',
      description='Library for reading and writing proio files and streams',
      url='http://github.com/decibelcooper/proio',
      author='David Blyth',
      author_email='dblyth@anl.gov',
      license='None',
      packages=['proio', 'proio.proto'],
      py_modules=modules,
      install_requires=['protobuf', 'lz4'],
      zip_safe=True)
