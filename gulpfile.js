const child = require('child_process');
const gulp = require('gulp');
const less = require('gulp-less');
const path = require('path');
const sequence = require('run-sequence');
const util = require('gulp-util');

var server;

gulp.task('default', ['go', 'less', 'watch']);
gulp.task('watch', ['watch:less', 'watch:go']);

gulp.task('go', function(callback) { sequence('go:build', 'go:start', callback); });

gulp.task('go:build', function() {
  if (server) {
    server.kill();
  }
  var build = child.spawnSync('go', ['install']);
  if (build.stderr.length) {
    var lines = build.stderr.toString()
      .split('\n').filter(function(line) {
        return line.length
      });
    for (var l in lines)
      util.log(util.colors.red(
        'Error (go install): ' + lines[l]
      ));
  }
  return build;
});

gulp.task('go:start', function() {
  if (server) {
    server.kill();
  }
  server = child.spawn(process.env.GOPATH + '/bin/S2017-UPE-AI.exe');
  /* Pretty print server log output */
  server.stdout.on('data', function(data) {
    var lines = data.toString().split('\n')
    for (var l in lines)
      if (lines[l].length)
        util.log(lines[l]);
  });
  /* Print errors to stdout */
  server.stderr.on('data', function(data) {
    process.stdout.write(data.toString());
  });
});

gulp.task('watch:go', function() {
  gulp.watch(['./**/*.go', './templates/*.html'], ['go:build', 'go:start']);
});

gulp.task('watch:less', function() {
  gulp.watch(['./static/less/*.less'], ['less']);
});

gulp.task('less', function() {
  return gulp.src('./static/less/*.less')
    .pipe(less({
      paths: [ path.join(__dirname, 'less', 'includes') ]
    }))
    .pipe(gulp.dest('./static/css'));
});
