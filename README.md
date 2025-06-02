# go-limit
workload for POC purpose

# use case
check if a limit was breach

# tables



   table spend_limit
   id|category    |mcc |day|hour|minute|amount |created_at                   |
   --+------------+----+---+----+------+-------+-----------------------------+
   1|CREDIT      |FOOD| 10| 100|    10|1000.00|2025-04-07 14:45:11.641 -0300|
   2|DEBIT       |FOOD| 10| 100|    10|1000.00|2025-04-07 14:45:11.641 -0300|
   3|CREDIT:TOKEN|FOOD| 10| 100|    10|1000.00|2025-04-07 14:45:11.641 -0300|

   table transaction_limit
   id    |transaction_id                      |category|card_number    |mcc |status   |transaction_at               |currency|amount|
   -----+------------------------------------+--------+---------------+----+---------+-----------------------------+--------+------+
   20525|50f6f552-9af3-48bc-8dba-ee3fffc09026|CREDIT  |111.111.111.500|FOOD|REQUESTED|2025-05-30 15:33:10.091 -0300|BRL     | 22.00|
   20524|c0c1e121-1fc3-424c-bbe4-2fc031333527|CREDIT  |111.111.111.500|FOOD|REQUESTED|2025-05-30 15:33:08.121 -0300|BRL     | 22.00|
   20523|e4605c5f-f7fa-460b-b4b8-f1fb1e83ce21|CREDIT  |111.111.111.500|FOOD|REQUESTED|2025-05-30 15:33:06.442 -0300|BRL     | 22.00|
   20522|a33f14bd-16e8-49e9-9a1f-6d52a2da869f|CREDIT  |111.111.111.500|FOOD|REQUESTED|2025-05-30 15:33:04.267 -0300|BRL     | 22.00|
   20521|7b3c213e-413f-4439-958f-0d75e474875b|CREDIT  |111.111.111.500|FOOD|REQUESTED|2025-05-30 15:33:00.920 -0300|BRL     | 22.00|
   20520|8e1ac718-49b6-47cd-ab34-b44d0fa632d9|CREDIT  |111.111.111.500|FOOD|REQUESTED|2025-05-30 15:32:59.107 -0300|BRL     | 22.00|
   20519|6f459405-b82b-45a3-99d8-3d80fcba6d2e|CREDIT  |111.111.111.500|FOOD|REQUESTED|2025-05-30 15:32:56.313 -0300|BRL     | 22.00|

   table breach_limit
   id  |fk_id_trans_limit|transaction_id                      |mcc |status             |amount   |count|created_at                   |
   ----+-----------------+------------------------------------+----+-------------------+---------+-----+-----------------------------+
      1|                2|1111-222-3333                       |FOOD|BREACH_LIMIT:CREDIT|  -100.09|    1|2025-04-07 14:33:43.329 -0300|
      2|                3|1111-222-3333                       |FOOD|BREACH_LIMIT:CREDIT|  -200.18|    2|2025-04-07 14:33:45.896 -0300|
      3|               11|1111-222-3333                       |FOOD|BREACH_LIMIT:CREDIT|    -0.90|   10|2025-04-07 14:55:20.540 -0300|
      4|               19|1111-222-3333                       |FOOD|BREACH_LIMIT:CREDIT|  -300.63|    7|2025-04-07 21:22:30.899 -0300|
      5|               20|1111-222-3333                       |FOOD|BREACH_LIMIT:CREDIT|  -700.72|    8|2025-04-07 21:22:32.649 -0300|
      6|               98|b6b8188d-7fcf-407b-8a97-af5f79fe4878|FOOD|BREACH_LIMIT:CREDIT|   -50.00|    8|2025-04-20 22:58:51.738 -0300|
      7|               99|413b6a08-eef8-48f9-9820-f1d299418ba6|FOOD|BREACH_LIMIT:CREDIT|  -150.00|    9|2025-04-20 23:00:18.144 -0300|