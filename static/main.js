import Athlete from "./Athlete.js";
import QueryForm from "./QueryForm.js";
import ActivityList from "./ActivityList.js";
import Lmap from "./Lmap.js";

const MultiActivityMapper = {
  data() {
    return {
      activityTypes: {},
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
    getActivities: function (data) {
      var $c = this;
      var text = encodeURIComponent(data.queryString);
      var p = fetch(
        `./activities?q=${text}&after=${data.fromDate}&before=${data.toDate}&type=${data.selectedType}`
      ).then((stream) => stream.json());
      p.then(this.$refs.map.addActivityData);
      p.then(function (data) {
        L.geoJSON(data).eachLayer($c.addToActivityList);
      });
    },

    addToActivityList: function (layer) {
      if (Object.prototype.hasOwnProperty.call(layer, "feature")) {
        var a = layer.feature.properties.activity;
        a.url = "https://strava.com/activities/" + a.id;
        if (!this.activities.find((b) => b.id == a.id)) {
          this.activities.push(a);
        }
      }
    },
  },
  mounted() {
    fetch(`./athlete`)
      .then((stream) => stream.json())
      .then((data) => (this.athlete = data))
      .catch(() => (this.athlete = null));
    fetch("./static/activitytypes.json")
      .then((stream) => stream.json())
      .then((data) => (this.activityTypes = data));
  },
};

Vue.createApp(MultiActivityMapper).mount("#app");
