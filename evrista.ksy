meta:
  id: evrista
  file-extension: gnt
  endian: le
  license: CC0-1.0

seq:
  - id: file
    type: file_desc
  - id: series
    type: series_desc
    repeat: expr
    repeat-expr: file.series_count
  - id: data
    type: series_data(_index)
    repeat: expr
    repeat-expr: file.series_count

types:

  file_desc:
    seq:
      - id: comment
        type: strz
        encoding: ascii
        size: 52
      - id: series_count
        type: u1
      - size: 16
        # TODO: unrecognized
      
  series_desc:
    seq:
    - id: name
      type: strz
      encoding: ascii
      size: 11
    - id: type
      type: u1
      enum: data_type
    - size: 2
      # TODO: always 0A 00
    - id: size
      type: u4
    - size: 87
      # TODO: unrecognized, contains references to other series by name
    
  series_data:
    params:
    - id: i
      type: u4
    seq:
    - id: values
      type:
        switch-on: _parent.series[i].type
        cases:
          data_type::float: f4
          data_type::double: f8
      repeat: expr
      repeat-expr: _parent.series[i].size

enums:
  data_type:
    0x05: float
    0x06: double
      