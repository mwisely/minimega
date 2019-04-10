/* global _ */

(function() {
  // DandelionGraph contains a MmGraph that draws a Host-Centric (?)
  // representation of a network. Both VLANs and Hosts are represented
  // as nodes. Buttons are included to recenter the graph and
  // expand/reduce the size of the nodes in the drawing.
  //
  //
  // Examples:
  //
  //     <dandelion-graph></dandelion-graph>
  //

  const template = `
      <div>
          <div class="btn-toolbar">
              <div class="btn-group">
                  <button class="btn btn-default" v-on:click="recenter()">
                      <i class="fa fa-repeat"></i>
                  </button>
              </div>
              <div class="btn-group pull-right">
                  <button
                      class="btn btn-default"
                      v-on:click="nodeRadius < 15 ? nodeRadius++ : nodeRadius">
                          <i class="fa fa-expand"></i>
                  </button>
                  <button
                      class="btn btn-default"
                      v-on:click="nodeRadius > 3 ? nodeRadius-- : nodeRadius">
                          <i class="fa fa-compress"></i>
                  </button>
              </div>
          </div>
          <mm-canvas>
              <mm-graph
               ref="graph"
               :nodes="nodes"
               :links="links"
               >
             </mm-graph>
          </mm-canvas>
      </div>
      `;

  window.DandelionGraph = {
    template: template,

    // Other components used by this Vue template
    components: {
      MmCanvas,
      MmGraph,
    },

    // Computed values are recomputed whenever dependencies change. If
    // dependencies don't change, the cached return value is used.
    computed: {
      // Returns an Array of nodes to be drawn. That is, an array of
      // Objects representing Hosts and VLANs is returned with relevant style
      // information. See the template and MmGraph for more details.
      nodes() {
        // VMs are blue
        const vms = _.map(this.$store.state.vms, (vm) => {
          return {
            id: 'vm' + vm.id,  // Prefix with 'vm' to avoid collisions w/ VLANs
            radius: vm.tags.hasOwnProperty('size') ? vm.tags.size : this.nodeRadius,
            fillStyle: vm.tags.hasOwnProperty('color') ? vm.tags.color : 'blue',
          };
        });

        // VLANs are red
        const vlans = _.map(this.$store.getters.vlans, (vlan) => {
          /*var colorMap = new Map();
          for ( i= 0; i < vlan.vms.length;i++){
            if (vlan.vms[i].tags.hasOwnProperty(vlan.name)){
              if (colorMap.has(vlan.vms[i].tags.color)){
                colorMap.set(vlan.vms[i].tags.color,colorMap.get(vlan.vms[i].tags.color)+1)
                continue
              }
              colorMap.set(vlan.vms[i].tags.color,1)
            }
          }*/
          return {
            id: 'vlan' + vlan.name,  // Prefix with 'vlan' to avoid collisions w/ VMs
            radius: 2,
            fillStyle: 'grey',
          };
        });

        // Glue 'em together
        return [].concat(vms, vlans);
      },

      // Returns an Array of links to be drawn. That is, an array of
      // Objects representing links between Hosts and VLANs is returned
      links() {
        const m = _.map(this.$store.state.vms, (vm) => {
          return _.map(vm.vlan, (vlan) => {
            return {
              source: 'vm' + vm.id,
              target: 'vlan' + vlan,
              strokeStyle: vm.tags.hasOwnProperty(vlan) ? vm.tags.color : 'black',
            };
          });
        });

        return _.flatten(m);
      },
    },

    // Local data for DandelionGraph
    data() {
      return {
        // The radius of the nodes
        nodeRadius: 5,
      };
    },

    // Convenience methods
    methods: {
      // Recenters and reheats the graph
      recenter() {
        this.$refs['graph'].recenter();
      },
    },
  };
})();
