// Time-stamp: <2024-09-13 19:42:14 krylon>
// -*- mode: javascript; coding: utf-8; -*-
// Copyright 2015-2020 Benjamin Walkenhorst <krylon@gmx.net>
//
// This file has grown quite a bit larger than I had anticipated.
// It is not a /big/ problem right now, but in the long run, I will have to
// break this thing up into several smaller files.

'use strict'

function defined (x) {
    return undefined !== x && null !== x
}

function fmtDateNumber (n) {
    return (n < 10 ? '0' : '') + n.toString()
} // function fmtDateNumber(n)

function timeStampString (t) {
    if ((typeof t) === 'string') {
        return t
    }

    const year = t.getYear() + 1900
    const month = fmtDateNumber(t.getMonth() + 1)
    const day = fmtDateNumber(t.getDate())
    const hour = fmtDateNumber(t.getHours())
    const minute = fmtDateNumber(t.getMinutes())
    const second = fmtDateNumber(t.getSeconds())

    const s =
          year + '-' + month + '-' + day +
          ' ' + hour + ':' + minute + ':' + second
    return s
} // function timeStampString(t)

function fmtDuration (seconds) {
    let minutes = 0
    let hours = 0

    while (seconds > 3599) {
        hours++
        seconds -= 3600
    }

    while (seconds > 59) {
        minutes++
        seconds -= 60
    }

    if (hours > 0) {
        return `${hours}h${minutes}m${seconds}s`
    } else if (minutes > 0) {
        return `${minutes}m${seconds}s`
    } else {
        return `${seconds}s`
    }
} // function fmtDuration(seconds)

function beaconLoop () {
    try {
        if (settings.beacon.active) {
            const req = $.get('/ajax/beacon',
                              {},
                              function (response) {
                                  let status = ''

                                  if (response.Status) {
                                      status = 
                                          response.Message +
                                          ' running on ' +
                                          response.Hostname +
                                          ' is alive at ' +
                                          response.Timestamp
                                  } else {
                                      status = 'Server is not responding'
                                  }

                                  const beaconDiv = $('#beacon')[0]

                                  if (defined(beaconDiv)) {
                                      beaconDiv.innerHTML = status
                                      beaconDiv.classList.remove('error')
                                  } else {
                                      console.log('Beacon field was not found')
                                  }
                              },
                              'json'
                             ).fail(function () {
                                 const beaconDiv = $('#beacon')[0]
                                 beaconDiv.innerHTML = 'Server is not responding'
                                 beaconDiv.classList.add('error')
                                 // logMsg("ERROR", "Server is not responding");
                             })
        }
    } finally {
        window.setTimeout(beaconLoop, settings.beacon.interval)
    }
} // function beaconLoop()

function beaconToggle () {
    settings.beacon.active = !settings.beacon.active
    saveSetting('beacon', 'active', settings.beacon.active)

    if (!settings.beacon.active) {
        const beaconDiv = $('#beacon')[0]
        beaconDiv.innerHTML = 'Beacon is suspended'
        beaconDiv.classList.remove('error')
    }
} // function beaconToggle()

/*
  The ‘content’ attribute of Window objects is deprecated.  Please use ‘window.top’ instead. interact.js:125:8
  Ignoring get or set of property that has [LenientThis] because the “this” object is incorrect. interact.js:125:8

*/

function db_maintenance () {
    const maintURL = '/ajax/db_maint'

    const req = $.get(
        maintURL,
        {},
        function (res) {
            if (!res.Status) {
                console.log(res.Message)
                postMessage(new Date(), 'ERROR', res.Message)
            } else {
                const msg = 'Database Maintenance performed without errors'
                console.log(msg)
                postMessage(new Date(), 'INFO', msg)
            }
        },
        'json'
    ).fail(function () {
        const msg = 'Error performing DB maintenance'
        console.log(msg)
        postMessage(new Date(), 'ERROR', msg)
    })
} // function db_maintenance()

function toggle_visibility(hostID) {
    const query = `tr.Host${hostID}`
    const visible = !$(`#show_${hostID}`)[0].checked
    $(query).each(function () {
        if (visible) {
            $(this).hide()
        } else {
            $(this).show()
        }
    })
} // function toggle_visibility(hostID)

function filter_source(src) {
    const query = `tr.src_${src}`
    const visible = !$(`#filter_src_${src}`)[0].checked
    $(query).each(function () {
        if (visible) {
            $(this).hide()
        } else {
            $(this).show()
        }
    })
} // function filter_source(src)

function toggle_visibility_all() {
    const visible = !$("#filter_toggle_all")[0].checked

    const boxes = jQuery("#sources_list input.filter_src_check")

    for (let i = 0; i < boxes.length; i++) {
        boxes[i].checked = !visible
    }

    const rows = jQuery("#records > tr")

    for (let i = 0; i < rows.length; i++) {
        rows[i].hide()
    }
} // function toggle_visibility_all()

function search_load_results(sid, page) {
    const addr = `/ajax/search/load/${sid}/${page}`

    const reply = $.get(addr,
                        {},
                        function (res) {
                            if (res.Status) {
                                jQuery("#results")[0].innerHTML = res.Payload["results"]

                                const params = JSON.parse(res.Payload["search"])

                                for (var h of params.Query.hosts) {
                                    const filter_id = `#filter_host_${h}`
                                    jQuery(filter_id)[0].checked = true
                                }

                                for (var s of params.Query.sources) {
                                    const filter_id = `#filter_src_${s}`
                                    jQuery(filter_id)[0].selected = true
                                }

                                if (params.Query.period.length == 2) {
                                    jQuery("#filter_period_begin")[0].valueAsDate =
                                        new Date(params.Query.period[0])
                                    jQuery("#filter_period_end")[0].valueAsDate =
                                        new Date(params.Query.period[1])
                                    jQuery("#filter_by_period_p")[0].checked = true
                                }

                                if (_.all(params.Query.terms, (x) => { return x.indexOf("(?i)") >= 0 })) {
                                    params.Query.terms =
                                        _.map(params.Query.terms, (x) => { return x.substring(4) })
                                    jQuery("#case_insensitive")[0].checked = true
                                } else {
                                    jQuery("#case_insensitive")[0].checked = false
                                }

                                jQuery("#search_terms")[0].value = params.Query.terms.join("\n")
                                jQuery("#search_id")[0].value = sid
                            } else {
                                const msg = `Error loading search results: ${res.Message}`
                                console.log(msg)
                                alert(msg)
                            }
                        },
                        'json'
                       ).fail((reply, status_text, xhr) => {
                           console.log(`Error searching: ${status_text} ${reply} ${xhr}`)
                       })
} // function search_load_results(sid, page)

function search_delete(id) {
    const addr = `/ajax/search/delete/${id}`

    const reply = $.get(addr,
                        {},
                        (res) => {
                            if (res.Status) {
                                const lid = `#search_${id}`
                                jQuery(lid)[0].remove()

                                const cur_id = Number.parseInt(jQuery("#search_id")[0].value)

                                if (id == cur_id) {
                                    clear_results()
                                    clear_filters()
                                    jQuery("#search_id")[0].value = ""
                                }
                            } else {
                                console.log(res.Message)
                                alert(res.Message)
                            }
                        },
                        'json')

    reply.fail((reply, status_text, xhr) => {
                           console.log(`Error searching: ${status_text} ${reply} ${xhr}`)
                       })
} // function search_delete(id)
