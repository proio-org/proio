from setuptools import setup

setup(name='proio',
      version='0.2.1',
      description='Library for reading and writing proio files and streams',
      url='http://github.com/decibelcooper/proio',
      author='David Blyth',
      author_email='dblyth@anl.gov',
      license='None',
      packages=['proio','proio.model','proio.model.lcio','proio.model.promc'],
      install_requires=['protobuf'],
      zip_safe=True)
