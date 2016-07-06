# Platypus

REST API for the Hunter Douglas Platinum Gateway. See [libhdplatinum](https://github.com/vincer/libhdplatinum)
for more details on features supported.

(I use REST a little liberally here. Note HATEOAS here. Also, I take some other liberties for the sake
of simplicity and easier client integration.)

## Routes

### GET /shades
Returns a list of shades.

```
[{
    "height": 0,
    "id": "00",
    "name": "Foo",
    "roomId": "00"
},
{
    "height": 50,
    "id": "01",
    "name": "Bar",
    "roomId": "01"
}
]
```

### GET /shades/:id
Returns a single shade. Example:
```
{
    "height": 0,
    "id": "00",
    "name": "Foo",
    "roomId": "00"
}
```

### PUT /shades/:id
Allows editing a shade. Expects the full `shade` object.

Example:
```
PUT /shades/00
{
    "height": 100,
    "id": "00",
    "name": "Foo",
    "roomId": "00"
}
```

Note: only height is currently editable. i.e. you can move your shades up and down.

### PUT /shades/:id/height
Allows editing only the height of the shade. For integration convenience.

Example:
```
PUT /shades/00/height
{
    "height": 100
}
```

## Errors

You can expect 400 when submitting obviously invalid data and 404 on missing resources. Otherwise
there isn't a whole lot of error checking, so you can likely expect some 500s.

## Height Range

Unline `libhdplatinum`, `platypus` uses a height range normalized to 0-100 (100 is fully open).
