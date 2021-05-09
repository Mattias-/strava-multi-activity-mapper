import Athlete from "./Athlete.js";
import QueryForm from "./QueryForm.js";
import ActivityList from "./ActivityList.js";
import Lmap from "./Lmap.js";

const MultiActivityMapper = {
  data() {
    return {
      authed: false,
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
  methods: {
    getActivities: function (data) {
      var text = encodeURIComponent(data.queryString);
      fetch(
        `./activities?q=${text}&after=${data.fromDate}&before=${data.toDate}&type=${data.selectedType}`
      )
        .then((stream) => stream.json())
        .then(this.$refs.map.addActivityData);
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
    var v = this;
    fetch(`./athlete`)
      .then((stream) => stream.json())
      .then(function (data) {
        var name = data.firstname + " " + data.lastname;
        v.athlete.name = name;
        v.athlete.url = "https://www.strava.com/athletes/" + data.id;
        v.athlete.image = data.profile_medium;
        v.authed = true;
      });
    fetch("./static/activitytypes.json")
      .then((stream) => stream.json())
      .then((data) => (this.activityTypes = data));
  },
};

Vue.createApp(MultiActivityMapper).mount("#app");
