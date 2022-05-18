
.PHONY: build backend frontend babel tsc

tsc:
	cd frontend; npm run check-types -- --watch

babel:
	cd frontend; npx babel --watch js_in/ --out-dir js_out/ --presets react-app/prod --extensions ".ts"

backend:
	cd backend; go run . 

frontend:
	cd frontend; npm run start

build:
	cd backend; go build
	
