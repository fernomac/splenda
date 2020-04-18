
const game = {
  props: {
    'game': Object,
  },
  methods: {
    'play': function() {
      window.location = '/games/'+this.game.id
    },
    'showMenu': function() {
      this.$emit('delete-menu', this.game.id)
    },
  },
  template: `
    <tr>
      <td>
        <div class="flex-column">
          <input type="button" class="button" value="play" @click="play">
          <input type="button" class="button" value="delete" @click="showMenu">
        </div>
      </td>
      <td>
        {{game.players.join(', ')}}
      </td>
      <td></td>
    </tr>
  `
}

const newmenu = {
  props: {
    'users': Array,
  },
  data: function() { return {
    selected: [],
  }},
  methods: {
    hide: function() {
      this.$emit('hide')
    },
    newGame: function() {
      fetch('/api/games', {
        method: 'POST',
        body: JSON.stringify({'players': this.selected})
      }).then(function(res) {
        if (res.ok) {
          res.json().then(function(json) {
            window.location = '/games/' + json.id
          })
        } else {
          res.text().then(function(err) {
            alert(err)
          })
        }
      })
    },
  },
  template: `
    <div class="menu">
      <div class="flex-row-between">
        <div style="opacity: 0;"><input type="button" class="button" value="X"></div>
        <div style="font-weight: bold; text-align: center;">Select Players</div>
        <div><input type="button" class="button" value="X" @click="hide"></div>
      </div>
      <div style="margin-top: 1em; margin-left: 2em;">
        <div v-for="user in users">
          <input type="checkbox" v-model="selected" :id="user" :value="user">
          <label :for="user">{{user}}</label>
        </div>
      </div>
      <div style="margin-top: 1em; text-align: center;">
        <input type="button" class="button" value="begin" @click="newGame">
      </div>
    </div>
  ` 
}

const deletemenu = {
  props: {
    'id': String,
  },
  methods: {
    hide: function() {
      this.$emit('hide')
    },
    deleteGame: function() {
      const menu = this
      fetch('/api/games/'+this.id, {
        method: 'DELETE',
      }).then(function(res) {
        if (res.ok) {
          menu.hide()
          update(app)
        } else {
          res.text().then(function(err) {
            alert(err)
          })
        }
      })
    }
  },
  template: `
    <div class="menu">
      <div style="display: flex; flex-direction: row; justify-content: flex-end;">
        <input type="button" class="button" value="X" @click="hide">
      </div>
      <div style="font-weight: bold; text-align: center;">Confirm?</div>
      <div style="margin-top: 1em; text-align: center;">
        <input type="button" class="button" value="delete" @click="deleteGame">
      </div>
    </div>
  `
}

function update(app) {
  fetch('/api/games').then(function(res) {
    if (res.ok) {
      res.json().then(function(json) {
        app.games = json.games
      })
    } else {
      res.text().then(function(err) {
        alert(err)
      })
    }
  })
}

const app = new Vue({
  el: '#root',
  data: {
    users: [],
    games: [],
    newMenu: false,
    deleteMenu: false,
    deleteID: '',
  },
  components: {
    'game': game,
    'newmenu': newmenu,
    'deletemenu': deletemenu,
  },
  methods: {
    showNewMenu: function() {
      const app = this
      fetch('/api/users').then(function(res) {
        if (res.ok) {
          res.json().then(function(json) {
            app.users = json.users
            app.newMenu = true
          })
        } else {
          res.text().then(function(err) {
            alert(err)
          })
        }
      })
    },
    hideNewMenu: function() {
      this.newMenu = false
    },
    showDeleteMenu: function(id) {
      this.deleteID = id
      this.deleteMenu = true
    },
    hideDeleteMenu: function() {
      this.deleteID = ''
      this.deleteMenu = false
    },
  },
  created: function() {
    update(this)
  }
})
