from setuptools import setup

setup(name='proio',
      version='0.2',
      description='Library for reading and writing proio files and streams',
      url='http://github.com/decibelcooper/proio',
      author='David Blyth',
      author_email='dblyth@anl.gov',
      license='None',
      packages=['proio','proio.model','proio.model.lcio','proio.model.promc'],
      zip_safe=True)
