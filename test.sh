
echo "=> clearing database ..."
curl http://127.0.0.1:8080/reset

echo "=> set initial players balances ..."
curl "http://127.0.0.1:8080/fund?playerId=P1&points=300"
curl "http://127.0.0.1:8080/fund?playerId=P2&points=300" 
curl "http://127.0.0.1:8080/fund?playerId=P3&points=300" 
curl "http://127.0.0.1:8080/fund?playerId=P4&points=500" 
curl "http://127.0.0.1:8080/fund?playerId=P5&points=1000" 

echo "=> setup tournament ..."
curl "http://127.0.0.1:8080/announceTournament?tournamentId=1&deposit=1000"
curl "http://127.0.0.1:8080/joinTournament?tournamentId=1&playerId=P5"
curl "http://127.0.0.1:8080/joinTournament?tournamentId=1&playerId=P1&backerId=P2&backerId=P3&backerId=P4"


echo "=> finish tournament ..."
curl -H "Content-Type: application/json" -X POST -d '{"tournamentId":"1","winners": [{"playerId": "P1", "prize": 2000}]}' http://127.0.0.1:8080/resultTournament

echo "=> check balances"
curl "http://127.0.0.1:8080/balance?playerId=P1"
curl "http://127.0.0.1:8080/balance?playerId=P2"
curl "http://127.0.0.1:8080/balance?playerId=P3"
curl "http://127.0.0.1:8080/balance?playerId=P4"
curl "http://127.0.0.1:8080/balance?playerId=P5"
