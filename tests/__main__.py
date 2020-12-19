# pylint: disable=missing-docstring

import os
import unittest

from . import db
from . import settings

def main():
    db.setup()
    path = settings.BASEPATH
    top_level_dir = os.path.dirname(path)
    tests = unittest.TestLoader().discover(path, top_level_dir=top_level_dir, pattern='test_*.py')
    unittest.TextTestRunner(verbosity=2).run(tests)

main()
