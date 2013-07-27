#!/usr/bin/env python

from __future__ import division
import os,sys,time
import pprint

# controller numbers
# brown's piano
#   pedals
#       67, 66, 64
#       value is binary: 127 when pedal is down, 0 when pedal is up
#       example: 176 64 127 (controller channel 0, # 64, value = 127)
#   instrument changes send this sequence of messages
#       channel 0 program-change, instrument = 0 (piano 1 and 2), 5 (epiano 1), 4 (epiano 2)
#       channel 0 controller 91 = 12 (piano 1), 19 (piano 2), 25 (epiano 1 and 2)
#       240, 67, 16, 76, 2, 1, 0, 1, 17
#       247
#       240, 67, 16, 76, 2, 1, 90, 1
#       247
#       240, 67, 16, 76, 2, 1, 64, 0, 0
#       247
#       channel 0 controller 94 = 0 (piano 1 and 2), 20 (epiano 1), 25 (epiano 2)
#       240, 67, 16, 76, 8, 0, 17, 127 (system command 0)
#       247
#       240, 67, 115, 127, 37, 17, 0, 61, 76
#       247 (system command 7)
#
# typical note message
#       middle C is key 60
#       144 60 65 (middle-c on with velocity 65)
#       144 60 0 (middle-c on with velocity 0)
#



# http://www.midi.org/techspecs/midimessages.php

# everything on the right-hand side is 7 bits (0 to 127)
COMMANDS = {
    0x8: ['note-off',         ['key', 'velocity']],
    0x9: ['note-on',          ['key', 'velocity']],
    0xa: ['aftertouch',       ['key', 'touch']],
    0xb: ['controller',       ['controller', 'value']], # controllers 120-127 are special
    0xc: ['program-change',   ['instrument']],
    0xd: ['channel-pressure', ['pressure']],
    0xe: ['pitch-bend',       ['lsb', 'msb']], # these are combined into "value"
    0xf: ['system',           []], # varies
}

# interesting system messages
# channel   message     
# 1         0nnndddd    time code quarter frame. n = message type, d = values
# 2         0lll.. 0mmm... song position pointer in beats
# 8         none        timing clock.  happens 24 times per quarter note
# 10        none        start
# 12        none        stop

# midi time code quarter frame details:
# https://en.wikipedia.org/wiki/MIDI_timecode#Quarter-frame_messages

print
print '-----------------------------------------'
print
def yieldMidi():
    f = file('/dev/midi1')
    message = []
    while True:
        byte = ord(f.read(1))
        if byte >= 128:
            if message:
                yield message
            message = [byte]
        else:
            message.append(byte)

def yieldMidiDict():
    for msg in yieldMidi():
        d = {}
        channel = msg[0] & 15
        kind = msg[0] >> 4
        command, fields = COMMANDS[kind]
        d['command'] = command
        d['channel'] = channel
        d['raw'] = msg
        if len(fields) != len(msg)-1:
            print 'error: msg has different length than expected.'
            print '    msg: %s' % msg
            print '    expected fields: %s' % fields
            continue # skip
        for field,val in zip(fields,msg[1:]):
            d[field] = val
        if d['command'] == 'pitch-bend':
            d['value'] = d['lsb'] | d['msb'] >> 7
        yield d


for d in yieldMidiDict():
    # skip clock messages
    if d['raw'][0] in [248, 254]: continue

    print ' '.join([hex(byte) for byte in d['raw']]).replace('0x','')
    print pprint.pformat(d)
    print

