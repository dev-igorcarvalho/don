#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

function setup() {
  const targetDir = process.cwd();
  const sourceAssetsDir = path.join(__dirname, '../assets');
  const artefattiDir = path.join(targetDir, '.artefatti');

  const subDirs = [
    '',
    'agents',
    'artifacts',
    'pipelines',
    'rules'
  ];

  console.log('🇮🇹 Inizializzazione della struttura in corso...');

  // Create directories
  for (const sub of subDirs) {
    const dirPath = path.join(artefattiDir, sub);
    if (!fs.existsSync(dirPath)) {
      try {
        fs.mkdirSync(dirPath, { recursive: true });
        console.log(`Created folder: ${path.relative(targetDir, dirPath)}`);
      } catch (err) {
        console.error(`❌ Error creating folder ${dirPath}:`, err.message);
        process.exit(1);
      }
    } else {
      console.log(`Folder already exists, skipped: ${path.relative(targetDir, dirPath)}`);
    }
  }

  // Files to copy
  const filesToCopy = [
    { src: 'README.md', dest: 'README.md' },
    { src: 'rules/pipeline-workflow-rules.md', dest: 'rules/pipeline-workflow-rules.md' },
    { src: 'rules/workflow.md', dest: 'rules/workflow.md' }
  ];

  for (const file of filesToCopy) {
    const srcPath = path.join(sourceAssetsDir, file.src);
    const destPath = path.join(artefattiDir, file.dest);

    if (!fs.existsSync(srcPath)) {
      console.error(`❌ Source file not found in assets: ${file.src}`);
      continue;
    }

    if (!fs.existsSync(destPath)) {
      try {
        fs.copyFileSync(srcPath, destPath);
        console.log(`Copied file: ${path.relative(targetDir, destPath)}`);
      } catch (err) {
        console.error(`❌ Error copying ${file.src}:`, err.message);
      }
    } else {
      console.log(`File already exists in target, skipped: ${path.relative(targetDir, destPath)}`);
    }
  }

  // Copy workflow rules to .claude folder if it exists
  const claudeDir = path.join(targetDir, '.claude');
  if (fs.existsSync(claudeDir)) {
    const claudeRulesDir = path.join(claudeDir, 'rules');
    if (!fs.existsSync(claudeRulesDir)) {
      try {
        fs.mkdirSync(claudeRulesDir, { recursive: true });
        console.log(`Created folder: ${path.relative(targetDir, claudeRulesDir)}`);
      } catch (err) {
        console.error(`❌ Error creating folder ${claudeRulesDir}:`, err.message);
        process.exit(1);
      }
    }

    const workflowSrc = path.join(sourceAssetsDir, 'rules/workflow.md');
    const workflowDest = path.join(claudeRulesDir, 'pipeline-workflow-rules.md');

    if (!fs.existsSync(workflowSrc)) {
      console.error(`❌ Source file not found in assets: rules/pipeline-workflow-rules.md`);
    } else {
      try {
        fs.copyFileSync(workflowSrc, workflowDest);
        console.log(`Copied (and overwrote) file: ${path.relative(targetDir, workflowDest)}`);
      } catch (err) {
        console.error(`❌ Error copying rules/pipeline-workflow-rules.md to .claude:`, err.message);
      }
    }
  }

  console.log('\n🌹 Salutiamo il nostro nuovo amico!');
  console.log('La struttura `.artefatti` è stata creata con successo nel tuo territorio.');
  console.log('Abbiamo messo le regole al loro posto. Ricorda: una promessa fatta è un debito pagato.');
  console.log('Fai buon uso di questo potere... prima che ti facciamo un\'offerta che non potrai rifiutare.');
  console.log('Benvenuto nella Famiglia.\n');
}

setup();
