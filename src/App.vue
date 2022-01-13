<script>
import Athlete from "./components/Athlete.vue";
import QueryForm from "./components/QueryForm.vue";
import ActivityList from "./components/ActivityList.vue";
import Lmap from "./components/Lmap.vue";

import Dexie from 'dexie';
export const db = new Dexie('mamDB');
db.version(1).stores({
  activities: '&id, *text, type, start_date',
});

export default {
  data() {
    return {
      athlete: {},
      activities: [],
    };
  },
  components: {
    Lmap,
    Athlete,
    QueryForm,
    ActivityList,
  },
  computed: {
    authed() {
      return this.athlete != null;
    },
  },
  methods: {
    clearShownActivities: function () {
      this.activities.length = 0;
      var map = this.$refs.map.mapObject;
      map.eachLayer(function (layer) {
        if (Object.prototype.hasOwnProperty.call(layer, "feature")) {
          map.removeLayer(layer);
        }
      });
    },
    clearCachedActivities: function () {
      db.activities.clear();
    },
    getActivities: function (data) {
      var $c = this;

      if (!data.onlyCached) {
        var text = encodeURIComponent(data.queryString);
        var p = fetch(
          `./activities?q=${text}&after=${data.fromDate}&before=${data.toDate}&type=${data.selectedType}`
        ).then((stream) => stream.json());

        p.then(function (data) {
          data.features.forEach(function (feature){
            var a = feature.properties.activity;
            db.activities.put({
              id: String(a.id),
              type: a.type,
              start_date: new Date(a.start_date),
              text: a.name.split(' ').concat(a.description.split(' ')),
              feature: feature,
            });
          });
        });
      }

      db.activities.where('text').anyOfIgnoreCase(data.queryString.split(' ')).distinct().and((value) => {
        var lower = new Date(data.fromDate)
        lower.setHours(0, 0, 0)
        var upper = new Date(data.toDate)
        upper.setHours(23, 59, 59)
        return value.start_date >= lower && value.start_date <= upper;
      }).each(function (data) {
        console.log("DB Found:", data)
        var feature = data.feature;
        var a = feature.properties.activity;
        $c.addToActivityList(a);
        $c.$refs.map.addActivityData(feature);
      }).then(() => {
        $c.$refs.map.centerMap();
      });

    },

    addToActivityList: function (a) {
        if (!this.activities.find((b) => b.id == a.id)) {
          a.url = "https://strava.com/activities/" + a.id;
          this.activities.push(a);
        }
    },
  },
  mounted() {
    fetch(`./athlete`)
      .then((stream) => stream.json())
      .then((data) => (this.athlete = data))
      .catch(() => (this.athlete = null));
  },
};
</script>

<template>
  <div id="menu">
    <template v-if="authed">
      <athlete v-bind:athlete="athlete"></athlete>
      <query-form
        v-on:submit="getActivities"
        v-on:clear="clearShownActivities"
        v-on:clearCache="clearCachedActivities"
      ></query-form>
      <activity-list v-bind:activities="activities"></activity-list>
    </template>
    <template v-else>
      <a href="auth" class="menu-image"
        ><img src="/connectwith_strava.svg" alt="Connect with Strava"
      /></a>
    </template>
    <a href="https://www.strava.com" class="menu-image"
      ><img src="/poweredby_strava.svg" alt="Powered by Strava"
    /></a>
  </div>
  <Lmap ref="map"></Lmap>
</template>

<style>
#menu {
  width: 400px;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: auto;
  justify-content: space-between;
  align-items: center;
}

@media screen and (max-width: 700px) {
  #menu {
    height: 50%;
    width: 100%;
  }
}

a.menu-image img {
  width: 200px;
}
</style>
