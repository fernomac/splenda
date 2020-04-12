const othercolor = {
  'red': 'white',
  'blue': 'white',
  'green': 'white',
  'black': 'white',
  'white': 'black',
}

var game = {}
var selected = {}

function renderNobles(nobles) {
  var html = '<div class="nobles">'

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

  html += '</div>'
  return html
}

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

function showMenu(id, tier, index) {
  const menu = document.getElementById('menu')
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

function hideMenu() {
  const menu = document.getElementById('menu')
  menu.style.display = 'none'
}

function renderCards(tier, cards, player) {
  var html = '<div class="cards">'
  var i = 0
  for (card of cards) {
    const id = (tier == 0 ? player : tier) + '_' + i
    html += '<div id="' + id + '" class="card ' + card.color + 'card'
    if (userid === game.current && (tier > 0 || userid === player)) {
      html += ' buyable" '
      html += 'onclick="showMenu(\'' + id + '\', ' + tier + ', ' + i + ')'
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
  var html = '<div class="coins">'
  for (color of ['green', 'white', 'blue', 'black', 'red', 'wild']) {
    html += '<div class="coin ' + color + '"><div class="num">' + coins[color] + '</div></div>'
  }
  html += '</div>'
  return html
}

function renderTable(table) {
  var html = renderNobles(table.nobles)
  html += '<div>'
  html += renderCards(3, table.cards[2])
  html += renderCards(2, table.cards[1])
  html += renderCards(1, table.cards[0])
  html += '</div>'
  html += renderCoins(table.coins)

  document.getElementById('table').innerHTML = html
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
  html += '</div>'
  html += renderCoins(player.coins)
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
  fetch('/api/games/'+gameid).then(function(res) {
    return res.json()
  }).then(function(json) {
    game = json
    render()
  })
}

function reserve() {
  fetch('/api/games/'+gameid+'/reserve', {
    method: 'POST',
    body: JSON.stringify(selected),
  }).then(function(res) {
    if (res.ok) {
      hideMenu()
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
      hideMenu()
      update()
    } else {
      res.text().then(function(body) {
        alert(body)
      })
    }
  })
}

window.onload = update
