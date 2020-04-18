
const coins = {
  data: function() { return {
    colors: ['green', 'white', 'blue', 'black', 'red', 'wild'],
  }},
  props: {
    'coins': Object,
  },
  template: `
    <div class="flex-row-evenly">
      <div v-for="color in colors" class="coin" :class="color">
        <div class="num">{{coins[color] || 0}}</div>
      </div>
    </div>
  `
}

const pcards = {
  props: {
    'cards': Object,
  },
  data: function() { return {
    colors: ['green', 'white', 'blue', 'black', 'red'],
  }},
  methods: {
    count: function(cards, color) {
      if (cards && cards[color]) {
        return cards[color].length
      }
      return 0
    },
  },
  template: `
    <div class="flex-row-evenly">
      <div v-for="color in colors" class="pcard" :class="color">
        <div class="num">{{count(cards, color)}}</div>
      </div>
      <div class="pcard hidden"></div>
    </div>
  `
}

const card = {
  props: {
    'card': Object,
    'tier': Number,
    'index': Number,
    'offlimits': Boolean,
  },
  computed: {
    buyable: function() {
      if (userid !== app.game.current) {
        return false
      }
      if (this.offlimits) {
        return false
      }
      // TODO: Calculate if we can afford it?
      return true
    },
    classes: function() {
      var ret = {'card': true}
      ret[this.card.color+'card'] = true
      ret['buyable'] = this.buyable
      return ret
    },
  },
  methods: {
    'select': function() {
      if (this.buyable) {
        this.$emit('select', {
          'tier': this.tier,
          'index': this.index
        })
      }
    }
  },
  template: `
    <div :class="classes" @click="select">
      <div class="top">
        <div class="points">{{card.points}}</div>
        <div class="gem" :class="card.color"></div>
      </div>
      <div class="info">
        <div v-for="(count, color) in card.cost" class="cost" :class="color">
          {{count}}
        </div>
      </div>
    </div>
  `
}

const cards = {
  props: {
    'cards': Array,
    'tier': Number,
    'offlimits': Boolean,
  },
  components: {
    'card': card,
  },
  methods: {
    'select': function(card) {
      this.$emit('select', card)
    },
  },
  template: `
    <div class="flex-row-evenly">
      <card v-for="(card, index) in cards"
        :card="card"
        :tier="tier"
        :index="index"
        :offlimits="offlimits"
        :key="card.id"
        @select="select($event)">
      </card>
    </div>
  `
}

const player = {
  props: {
    'player': Object,
    'current': String,
  },
  components: {
    'coins': coins,
    'cards': pcards,
    'reserved': cards,
  },
  methods: {
    'select': function(card) {
      this.$emit('select', card)
    }
  },
  template: `
    <div class="player">
      <div class="header">
        <div class="id">{{player.id}} <span v-show="player.id === current"> 👈</span></div>
        <div class="points">points: {{player.points}}</div>
      </div>
      <coins :coins="player.coins"></coins>
      <cards :cards="player.cards"></cards>
      <reserved
        :cards="player.reserved"
        :tier="0"
        :offlimits="player.id !== current"
        @select="select($event)">
      </reserved>
    </div>
  `
}

const noble = {
  props: {
    'noble': Object,
  },
  template: `
    <div class="noble">
      <div class="info">
        <div class="points">{{noble.points}}</div>
        <div style="flex-grow: 1;"></div>
        <div v-for="(count, color) in noble.cost" class="cost" :class="color">
          {{count}}
        </div>
      </div>
    </div>
  `
}

const nobles = {
  props: {
    'nobles': Array,
  },
  components: {
    'noble': noble,
  },
  template: `
    <div class="flex-row-evenly">
      <noble v-for="noble in nobles"
        :noble="noble"
        :key="noble.id">
      </noble>
    </div>
  `
}

function count(els) {
  var count = 0
  for (var el of els) {
    if (el.checked) {
      count++
    }
  }
  return count
}

const takemenu = {
  props: {
    'title': String,
    'num': Number,
  },
  data: function() { return {
    'colors': [],
    'disabled': {
      'red': false,
      'blue': false,
      'green': false,
      'black': false,
      'white': false,
    },
    'takeable': false,
  }},
  methods: {
    'validate': function() {
      const enough = (this.colors.length === this.num)
      this.takeable = enough
      for (var color in this.disabled) {
        this.disabled[color] = (enough && !this.colors.includes(color))
      }
    },
    'take': function() {
      this.$emit('take', this.colors)
    },
    'cancel': function() {
      this.$emit('cancel')
    }
  },
  template: `
    <div>
      <div style="font-weight: bold; text-align: center;">{{title}}</div>
      <div style="height: 1em;"></div>

      <div>
        <input type="checkbox" id="green" value="green" v-model="colors" @change="validate" :disabled="disabled['green']">
        <label for="green" style="color: green;">green</label>
      </div>
      <div>
        <input type="checkbox" id="white" value="white" v-model="colors" @change="validate" :disabled="disabled['white']">
        <label for="white">white</label>
      </div>
      <div>
        <input type="checkbox" id="blue" value="blue" v-model="colors" @change="validate" :disabled="disabled['blue']">
        <label for="blue" style="color: blue;">blue</label>
      </div>
      <div>
        <input type="checkbox" id="black" value="black" v-model="colors" @change="validate" :disabled="disabled['black']">
        <label for="black">black</label>
      </div>
      <div>
        <input type="checkbox" id="red" value="red" v-model="colors" @change="validate" :disabled="disabled['red']">
        <label style="color: red;">red</label>
      </div>

      <div style="height: 1em;"></div>
      <input type="button" class="button" value="take" @click="take" :disabled="!takeable">
      <input type="button" class="button" value="cancel" @click="cancel">
    </div>
  `
}

function findplayer(id, players) {
  for (var player of players) {
    if (player.id === id) {
      return player
    }
  }
  return null
}

const buymenu = {
  props: {
    'title': String,
    'label': String,
    'game': Object,
    'selection': Object,
  },
  computed: {
    'card': function() {
      const tier = this.selection.tier
      const index = this.selection.index
      if (tier !== undefined && index !== undefined) {
        if (tier === 0) {
          if (this.label === 'reserve') {
            // HACK HACK HACK.
            return null
          }
          return findplayer(userid, this.game.players).reserved[index]
        }
        return this.game.table.cards[tier-1][index]
      }
      return null
    }
  },
  methods: {
    'buy': function() {
      this.$emit('buy', this.selection)
    },
    'cancel': function() {
      this.$emit('cancel')
    }
  },
  components: {
    'card': card,
  },
  template: `
    <div>
      <div style="font-weight: bold; text-align: center;">{{title}}</div>
      <div style="height: 1em;"></div>
      <div style="display: flex; flex-direction: column; align-items: center;">
        <card v-if="card" :card="card" :offlimits="true"></card>
      </div>
      <div style="height: 1em;"></div>
      <input type="button" class="button" :value="label" @click="buy" :disabled="card===null">
      <input type="button" class="button" value="cancel" @click="cancel">
    </div>
  `
}

const leftPane = {
  props: {
    'game': Object,
    'selection': Object,
  },
  data: function() { return {
    menu: '',
  }},
  watch: {
    'selection': function(newS, oldS) {
      // If there is no menu open, and the user selects a
      // card, open up the buy menu.
      if (this.menu === '' && newS.tier !== undefined) {
        this.menu = 'buy'
      }
    }
  },
  methods: {
    'finish': function() {
      this.menu = ''
      this.$emit('finished')
    },
    'handle': function(res) {
      if (res.ok) {
        this.finish()
        update(app)
      } else {
        res.text().then(function(body) {
          alert(body)
        })
      }
    },
    'take3': function(colors) {
      fetch('/api/games/'+gameid+'/take3', {
        method: 'POST',
        body: JSON.stringify({'colors': colors})
      }).then(this.handle)
    },
    'take2': function(colors) {
      fetch('/api/games/'+gameid+'/take2', {
        method: 'POST',
        body: JSON.stringify({'color': colors[0]})
      }).then(this.handle)
    },
    'reserve': function(card) {
      fetch('/api/games/'+gameid+'/reserve', {
        method: 'POST',
        body: JSON.stringify(card),
      }).then(this.handle)
    },
    'buy': function(card) {
      fetch('/api/games/'+gameid+'/buy', {
        method: 'POST',
        body: JSON.stringify(card),
      }).then(this.handle)
    },
  },
  components: {
    'takemenu': takemenu,
    'buymenu': buymenu,
  },
  template: `
    <div class="left-pane">
      <div style="height: 0.5em;"></div>
      <div class="flex-column">
        <input type="button" class="button" value="take 3 coins" @click="menu = 'take3'">
        <input type="button" class="button" value="take 2 coins" @click="menu = 'take2'">
        <input type="button" class="button" value="reserve card" @click="menu = 'reserve'">
        <input type="button" class="button" value="buy card" @click="menu = 'buy'">
      </div>

      <div style="height: 1em;"></div>

      <takemenu v-show="menu==='take3'" title="take 3 coins" :num="3" @take="take3($event)" @cancel="menu = ''"></takemenu>
      <takemenu v-show="menu==='take2'" title="take 2 coins" :num="1" @take="take2($event)" @cancel="menu = ''"></takemenu>
      <buymenu v-show="menu==='reserve'" title="reserve card" label="reserve" :game="game" :selection="selection" @buy="reserve($event)" @cancel="menu = ''"></buymenu>
      <buymenu v-show="menu==='buy'" title="buy card" label="buy" :game="game" :selection="selection" @buy="buy($event)" @cancel="menu = ''"></buymenu>
    </div>
  `
}

const centerPane = {
  props: {
    'table': Object,
  },
  components: {
    'nobles': nobles,
    'cards': cards,
    'coins': coins,
  },
  methods: {
    'select': function(card) {
      this.$emit('select', card)
    },
  },
  template: `
    <div class="center-pane">
      <div style="height: 1em;"></div>
      <nobles :nobles="table.nobles"></nobles>
      <div>
        <cards :cards="table.cards[2]" :tier="3" :offlimits="false" @select="select($event)"></cards>
        <cards :cards="table.cards[1]" :tier="2" :offlimits="false" @select="select($event)"></cards>
        <cards :cards="table.cards[0]" :tier="1" :offlimits="false" @select="select($event)"></cards>
      </div>
      <coins :coins="table.coins"></coins>
    </div>
  `
}

const rightPane = {
  props: {
    'players': Array,
    'current': String,
  },
  components: {
    'player': player,
  },
  methods: {
    'select': function(card) {
      this.$emit('select', card)
    }
  },
  template: `
    <div class="right-pane">
      <player v-for="player in players"
        :player="player"
        :current="current"
        :key="player.id"
        @select="select($event)">
      </player>
      <div>&nbsp;</div>
    </div>
  `
}

Vue.component('game', {
  props: {
    'game': Object,
  },
  data: function() { return {
    'selection': {},
  }},
  components: {
    'left-pane': leftPane,
    'center-pane': centerPane,
    'right-pane': rightPane,
  },
  methods: {
    'select': function(card) {
      this.selection = card
    },
    'unselect': function() {
      this.selection = {}
    },
  },
  template: `
    <div class="flex-row full-height">
      <left-pane :game="game" :selection="selection" @finished="unselect"></left-pane>
      <center-pane :table="game.table" @select="select($event)"></center-pane>
      <right-pane :players="game.players" :current="game.current" @select="select($event)"></right-pane>
    </div>
  `
})

function update(app) {
  fetch('/api/games/'+gameid).then(function(res) {
    if (res.ok) {
      res.json().then(function(json) {
        app.game = json

        if (json.current != userid) {
          setTimeout(function() {
            update(app)
          }, 1000)
        }
      })
    } else {
      res.text().then(function(text) {
        alert(text)
      })
    }
  })
}

const app = new Vue({
  el: '#root',
  data: {
    'game': {
      'table': {
        'nobles': [],
        'cards': [],
        'coins': {},
      },
    },
  },
  created: function() {
    update(this)
  },
})
