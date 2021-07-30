package main

import (
   "context"
   "fmt"
   "os"
   "strconv"
   "net/http"
   "github.com/jackc/pgx/v4/pgxpool"
   "github.com/gorilla/mux"
)

func fetchFibMemoCt(val int64, pg *pgxpool.Pool) string {
   var ct int
   err := pg.QueryRow(context.Background(), "select count(1) from fib_memo where val < $1 and val > 0", val).Scan(&ct)
   if err != nil {
      fmt.Println("error getting count of values", err)
      return "-1"
   } else {
      fmt.Println("found", ct, "memoized entries less than", val)
      return strconv.Itoa(ct)
   }
}

func purge(pg *pgxpool.Pool) bool {
   fmt.Println("processing request to purge all memoized entries")
   if _, err := pg.Exec(context.Background(), "TRUNCATE fib_memo")  ; err != nil {
      fmt.Println("error truncating table:", err)
      return false
   } else {
      fmt.Println("successfully truncated memoization table")
      storeFibVal(1, 1, pg)
      storeFibVal(2, 1, pg)
      return true
   }
}

func storeFibVal(n int64, val int64, pg *pgxpool.Pool) {
   fmt.Println("storing Fibonacci number", n, "value", val, "in database")
   _, err := pg.Exec(context.Background(), "INSERT INTO fib_memo (n, val)VALUES ($1, $2) ON CONFLICT (n) DO UPDATE SET n = excluded.n, val = excluded.val", n, val)
   if err != nil {
      fmt.Println("error inserting memo values for", n, "and", val, ":", err)
   }
}

func computeFib(n int64, pg *pgxpool.Pool) int64 {
   // zeroth: return the "silly" cases

   fmt.Println("processing request to compute Fibonacci number", n)

   switch {
      case n < 1:
         fmt.Println("this doesn't do negative Fibonacci; wait for 2.0")
         return -1
      case n == 1:
         return 1
      case n == 2:
         return 1
   }

   // first:  figure out if this IS one of the memoized values; if so, return that
   var fib int64
   err := pg.QueryRow(context.Background(), "select val from fib_memo where n = $1", n).Scan(&fib)
   if err == nil {
      fmt.Println("found memoized value", fib, "for Fibonacci number", n, "and returning it")
      return fib
   }

   // okay; that didn't work; so, now, obtain the largest memoized value, less than n, that has an n-1 value behind it
   // "but Theron, it will never be the case that there will be a value in this database that doesn't have
   // the immediately prior value also memoized"
   // HAHAHAHAHAHAHAHAHAHAHAHAHAHAHAHAAHAHAHAHAHAHAHAHAHAHA
   // *wipes tear from eye*
   // ahhh, good times
   // ... anyway
   // funky SQL query time - thanks for picking PostgreSQL, it has the functions I need

   var (
      working_n int64 = 3
      working_n_1 int64
      val int64 = 1
      val_1 int64 = 1
   )
   
   _ = pg.QueryRow(context.Background(), `SELECT n, val, n_1, n_1_val  
                                          FROM ( SELECT n, lead(n) over (order by n desc) as n_1, val, lead(val) over (order by n desc) as n_1_val  
                                                 FROM fib_memo  
                                                 WHERE n < $1 ) x  
                                          WHERE n_1 = n-1 limit 1`, n).Scan(&working_n, &val, &working_n_1, &val_1)

   // now we have our starting point - worst-case, we start from the beginning
  
   for ; working_n < n; {
      fib = val + val_1
      val_1 = val
      val = fib
      // fmt.Println("step values: ", working_n, fib, val, val_1)
      working_n++
      storeFibVal(working_n, fib, pg)
   }

   // ding!  done
   fmt.Println("computed Fibonacci number", n, "value:", fib)
   return fib
}

func main() {
   fmt.Println("starting")
   pg, err := pgxpool.Connect(context.Background(), "postgres://fib:fib@127.0.0.1:5432/fib")
   if err != nil {
      fmt.Println("cannot connect to database: ", err)
      os.Exit(1)
   }
   defer pg.Close()

   fmt.Println("connected to database, starting request router")

   // prime the pump
   purge(pg)


   r := mux.NewRouter()

   r.HandleFunc("/fib/{n}", func(w http.ResponseWriter, r *http.Request) {
                                    vars := mux.Vars(r)
                                    n, err := strconv.ParseInt(vars["n"], 10, 64) 
                                    if err == nil { 
                                       fmt.Println("computing Fibonacci for", n)
                                       x := computeFib(n, pg)
                                       if ( x == -1 ) {
                                          w.WriteHeader(404)
                                          w.Write([]byte("Error computing Fibonacci number"))
                                       } else {
                                          w.WriteHeader(http.StatusOK)
                                          fmt.Fprintf(w, strconv.FormatInt(x, 10))
                                       }
                                    } else {
                                       fmt.Println("error parsing request: ", err)
                                       w.WriteHeader(404)
                                       w.Write([]byte("Badly formed request"))
                                    }
                                 })
   r.HandleFunc("/fetchmemoct/{val}", func(w http.ResponseWriter, r *http.Request) {
                                    vars := mux.Vars(r)
                                    val, err := strconv.ParseInt(vars["val"], 10, 64) 
                                    if err == nil { 
                                       fmt.Println("getting number of memoized values below", val)
                                       w.WriteHeader(http.StatusOK)
                                       fmt.Fprintf(w, fetchFibMemoCt(val, pg))
                                    } else {
                                       fmt.Println("error parsing request: ", err)
                                       w.WriteHeader(404)
                                       w.Write([]byte("Badly formed request"))
                                    }
                                 })
   r.HandleFunc("/purge", func(w http.ResponseWriter, r *http.Request) {
                                    fmt.Println("received request to all memoized values")
                                    if purge(pg) {
                                       w.WriteHeader(http.StatusOK)
                                       fmt.Fprintf(w, "1")
                                    } else {
                                       fmt.Println("error purging memoized values")
                                       w.WriteHeader(404)
                                       w.Write([]byte("Error purging memoized values"))
                                    }
                               })

   fmt.Println("starting service")

   http.ListenAndServe(":80", r)
}
