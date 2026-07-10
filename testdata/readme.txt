go-omf test data directory.

OMF is an HDF5-based binary format. Small test OMF files
can be generated using the Python omf library:

    pip install omf
    python -c "
import omfvista
# Create a simple triangle mesh
mesh = omfvista.WarpByScalar(...)
omfvista.save('test.omf', mesh)
"

Currently go-omf does not have a HDF5-based reader.
The test files in this directory serve as placeholders for
future binary format testing.
