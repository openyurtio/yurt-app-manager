#!/usr/bin/env bash

set -ue

num_nodes=${1:-1}
tmp_fn_prefix=vk-cfg
vk_node_prefix=vkubelet-mock

template="{
  \"{{.node_prefix}}-{{.id}}\": {
    \"cpu\": \"2\",
    \"memory\": \"32Gi\",
    \"pods\": \"128\"
  }
}"

# generate configuration file
for i in $(seq 1 $num_nodes);
do
    echo "$template" | 
        sed "s|{{.id}}|$i|g; s|{{.node_prefix}}|$vk_node_prefix|g" \
        > $tmp_fn_prefix-$i
done

# add virtual node to the cluster
for i in $(seq 1 $num_nodes);
do 
    ./virtual-kubelet \
        --provider-config $tmp_fn_prefix-$i \
        --klog.v=5 --provider mock --nodename $vk_node_prefix-$i \
        --metrics-addr ":0" &
done

# remove all vk nodes and temporary configuration file
kill_all_vks() {
    for i in $(seq 1 $num_nodes)
    do
        rm $tmp_fn_prefix-$i
        kubectl delete node $vk_node_prefix-$i
    done
    exit
}

# capture the ctrl-c signal
trap kill_all_vks SIGINT

# block 
while :; do sleep 2073600; done
