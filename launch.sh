# sh launch.sh {3 4 5 6 7...}
# argument : number of nodes (containers)

if [ "$1" == '' ]; then
    echo "no first argument --> 3 containers is the default"
    NUM=3
else
    NUM=$1
fi

if [ $NUM -lt 2 ]; then
    echo first argument too small choose between 2 and 99
    exit 1
fi

if [ $NUM -gt 99 ]; then
    echo first argument too big choose between 2 and 99
    exit 1
fi

docker network create state-machine-network
echo "Building state machine server image"
docker build -t $"sm-server-app" ./state-machine-server/.

for i in $(seq 1 $NUM); do
    PORT=$((8000 + i))
    echo "Launching state machine server n°$i at port $PORT"
docker run -id --rm --name $"sm-server-$i" --net state-machine-network -p $PORT:8000 -d $"sm-server-app" ./state-machine $i $NUM
done
