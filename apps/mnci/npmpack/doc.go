// Package npmpack is `mnci npm-pack`, a step of the node plugin's package
// task: an app that builds into dist/<project dir> is packed from there, once
// its package.json is staged beside the build. Hidden: tasks call it through
// {mnci}, people call mnci ci pack.
package npmpack
