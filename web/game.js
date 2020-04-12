const othercolor = {
  'red': 'white',
  'blue': 'white',
  'green': 'white',
  'black': 'white',
  'white': 'black',
}

var game = {}
var selected = {}

function offset(el) {
  const left = window.pageXOffset || document.documentElement.scrollLeft
  const top = window.pageYOffset || document.documentElement.scrollTop
  const rect = el.getBoundingClientRect()
  return {
    left: left+rect.left,
    right: left+rect.right,
    top: top+rect.top,
    bottom: top+rect.bottom
  }
}

function showBuyMenu(id, tier, index) {
  hideTake2Menu()
  hideTake3Menu()

  const menu = document.getElementById('buy-menu')
  const rect = offset(document.getElementById(id))
  menu.style.top = rect.top - 15
  if (tier == 0) {
    menu.style.left = rect.left - 120
    document.getElementById('reserve').style.display = 'none'
  } else {
    menu.style.left = rect.right + 15
    document.getElementById('reserve').style.display = 'initial'
  }
  menu.style.display = 'flex'
  selected.tier = tier
  selected.index = index
}

function showTake3Menu() {
  hideBuyMenu()
  hideTake2Menu()

  for (el of document.getElementsByName('take3')) {
    el.disabled = false
    el.checked = false
  }

  const menu = document.getElementById('take3-menu')
  const rect = offset(document.getElementById('take3-button'))
  menu.style.top = rect.top - 150
  menu.style.left = rect.right + 15
  menu.style.display = 'flex'
}

function showTake2Menu() {
  hideBuyMenu()
  hideTake3Menu()

  for (el of document.getElementsByName('take2')) {
    el.disabled = false
    el.checked = false
  }

  const menu = document.getElementById('take2-menu')
  const rect = offset(document.getElementById('take2-button'))
  menu.style.top = rect.top - 150
  menu.style.left = rect.right + 15
  menu.style.display = 'flex'
}

function hideBuyMenu() {
  document.getElementById('buy-menu').style.display = 'none'
}

function hideTake3Menu() {
  document.getElementById('take3-menu').style.display = 'none'
}

function hideTake2Menu() {
  document.getElementById('take2-menu').style.display = 'none'
}

function threeChecked() {
  var count = 0
  for (el of document.getElementsByName('take3')) {
    if (el.checked) {
      count++
    }
  }
  return count >= 3
}

function validateTake3() {
  var enough = false
  if (threeChecked()) {
    enough = true
  }
  for (el of document.getElementsByName('take3')) {
    el.disabled = (enough && !el.checked)
  }
}

function oneChecked() {
  var count = 0
  for (el of document.getElementsByName('take2')) {
    if (el.checked) {
      count++
    }
  }
  return count >= 1
}

function validateTake2() {
  var enough = false
  if (oneChecked()) {
    enough = true
  }
  for (el of document.getElementsByName('take2')) {
    el.disabled = (enough && !el.checked)
  }
}

function renderNobles(nobles) {
  var html = ''

  for (noble of nobles) {
    html += '<div class="noble">'
    html += '<div class="info">'
    html += '<div class="points">' + noble.points + '</div>'
    html += '<div style="flex-grow: 1;"></div>'
    for (color in noble.cost) {
      html += '<div class="cost ' + color + '">'
      html += noble.cost[color]
      html += '</div>'
    }
    html += '</div></div>'
  }

  document.getElementById('nobles').innerHTML = html
}

function renderCards(tier, cards, player) {
  var html = '<div class="cards">'
  var i = 0
  for (card of cards) {
    const id = (tier == 0 ? player : tier) + '_' + i
    html += '<div id="' + id + '" class="card ' + card.color + 'card'
    if (userid === game.current && (tier > 0 || userid === player)) {
      html += ' buyable" '
      html += 'onclick="showBuyMenu(\'' + id + '\', ' + tier + ', ' + i + ')'
    }
    html += '">'
    html += '<div class="top">'
    html += '<div class="points">' + card.points + '</div>'
    html += '<div class="gem ' + card.color + '"></div>'
    html += '</div><div class="info">'
    for (color in card.cost) {
      html += '<div class="cost ' + color + '">'
      html += card.cost[color]
      html += '</div>'
    }
    html += '</div></div>'
    i++
  }
  html += '</div>'
  return html
}

function renderCoins(coins) {
  var html = ''
  for (color of ['green', 'white', 'blue', 'black', 'red', 'wild']) {
    html += '<div class="coin ' + color + '"><div class="num">' + coins[color] + '</div></div>'
  }
  return html
}

function renderTableCards(cards) {
  var html = ''
  for (i of [2, 1, 0]) {
    html += renderCards(i+1, cards[i])
  }
  document.getElementById('cards').innerHTML = html
}

function renderTableCoins(coins) {
  document.getElementById('coins').innerHTML = renderCoins(coins)
}

function renderTable(table) {
  renderNobles(table.nobles)
  renderTableCards(table.cards)
  renderTableCoins(table.coins)
}

function renderPlayerCards(cards) {
  var html = '<div class="cards">'
  for (color of ['green', 'white', 'blue', 'black', 'red']) {
    var count = 0
    if (cards[color]) {
      count = cards[color].length
    }
    html += '<div class="smallcard ' + color + '">'
    html += '<div class="num">' + count + '</div>'
    html += '</div>'
  }
  html += '<div style="width: 1.2em; border: 1px solid transparent;"></div>'
  html += '</div>'
  return html
}

function renderPlayer(player) {
  var html = '<div class="player">'
  html += '<div class="top">'
  html += '<div class="id">' + player.id
  if (player.id == game.current) {
    html += ' 👈'
  }
  html += '</div>'
  html += '<div class="points">score: ' + player.points + '</div>'
  html += '</div><div class="coins">'
  html += renderCoins(player.coins)
  html += '</div>'
  html += renderPlayerCards(player.cards)
  html += renderCards(0, player.reserved, player.id)
  html += '</div>'
  return html
}

function renderPlayers(players) {
  var html = ""
  for (player of players) {
    html += renderPlayer(player)
  }
  html += "<div>&nbsp;</div>"
  document.getElementById('players').innerHTML = html
}

function render() {
  renderTable(game.table)
  renderPlayers(game.players)
}

function update() {
  fetch('/api/games/'+gameid).then((res) => res.json()).then((json) => {
    game = json
    render()

    if (game.current != userid) {
      setTimeout(update, 1000)
    }
  })
}

function take3() {
  var colors = []
  for (el of document.getElementsByName('take3')) {
    if (el.checked) {
      colors.push(el.value)
    }
  }

  fetch('/api/games/'+gameid+'/take3', {
    method: 'POST',
    body: JSON.stringify({'colors': colors}),
  }).then((res) => {
    if (res.ok) {
      hideTake3Menu()
      update()
    } else {
      res.text().then(function(body) {
        alert(body)
      })
    }
  })
}

function take2() {
  var color = ''
  for (el of document.getElementsByName('take2')) {
    if (el.checked) {
      color = el.value
    }
  }

  fetch('/api/games/'+gameid+'/take2', {
    method: 'POST',
    body: JSON.stringify({'color': color}),
  }).then((res) => {
    if (res.ok) {
      hideTake2Menu()
      update()
    } else {
      res.text().then(function(body) {
        alert(body)
      })
    }
  })
}

function reserve() {
  fetch('/api/games/'+gameid+'/reserve', {
    method: 'POST',
    body: JSON.stringify(selected),
  }).then(function(res) {
    if (res.ok) {
      hideBuyMenu()
      update()
    } else {
      res.text().then(function(body) {
        alert(body)
      })
    }
  })
}

function buy() {
  fetch('/api/games/'+gameid+'/buy', {
    method: 'POST',
    body: JSON.stringify(selected),
  }).then(function(res) {
    if (res.ok) {
      hideBuyMenu()
      update()
    } else {
      res.text().then(function(body) {
        alert(body)
      })
    }
  })
}

window.onload = update
